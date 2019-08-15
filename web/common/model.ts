import * as PIXI from "pixi.js";
import { ResidenceContainer, CompanyContainer } from "./models/background";
import MonitorContainer from "./models/container";
import { Coordinates, config } from "./interfaces/gamemap";
import { Monitorable } from "./interfaces/monitor";
import { GameModelProperty, ResourceAttachable } from "./interfaces/pixi";
import { GameMap } from "../state";
import { RailEdge, RailNodeContainer, RailEdgeContainer, RailNode } from "./models/rail";
import { StationContainer } from "./models/station";
import CursorModel from "./models/cursor";
import { WorldBorder } from "./models/border";

const forceMove = { forceMove: true };

export default class implements ResourceAttachable {
    protected app: PIXI.Application;
    renderer: PIXI.Renderer;
    protected payload: {[index:string]: MonitorContainer<Monitorable>} = {};
    protected world: WorldBorder;
    protected changed: boolean = false;
    cursor: CursorModel;
    timestamp: number;
    textures: {[index: string]: PIXI.Texture};
    coord: Coordinates;
    offset: number;
    debugText: PIXI.Text;
    debugValue: any;

    constructor(options: GameModelProperty) {
        this.app = options.app;
        this.renderer = options.app.renderer;
        this.textures = {};
       
        this.coord = { cx: options.cx, cy: options.cy, scale: options.scale, zoom: options.zoom };
        this.timestamp = 0;
        this.offset = 0;

        this.cursor = new CursorModel({ app: this.app, offset: this.offset });
        this.cursor.setupDefaultValues();
        this.cursor.setupUpdateCallback();
        this.cursor.setupBeforeCallback();
        this.cursor.setupAfterCallback();
        this.cursor.setInitialValues({ visible: false, x: -1, y: -1 });
        this.cursor.begin();

        this.world = new WorldBorder({ app: this.app });
        this.world.setupDefaultValues();
        this.world.setupUpdateCallback();
        this.world.setupBeforeCallback();
        this.world.setupAfterCallback();
        this.world.setInitialValues({});
        this.world.begin();

        this.app.ticker.add(() => {
            this.offset++;
            if (this.offset >= config.round) {
                this.offset = 0;
            }
            Object.keys(this.payload).forEach(key => {
                this.payload[key].merge("offset", this.offset);
                this.payload[key].endChildren();
                this.cursor.merge("offset", this.offset);
            });
        });

        this.debugText = new PIXI.Text("");
        this.debugText.style.fontSize = 14;
        this.debugText.style.fill = 0xffffff;
        this.app.stage.addChild(this.debugText);
        setInterval(() => this.viewDebugInfo(), 250);
    }

    attach(textures: {[index: string]: PIXI.Texture}) {
        this.payload["residences"] = new ResidenceContainer({ app: this.app, texture: textures.residence});
        this.payload["companies"] = new CompanyContainer({ app: this.app, texture: textures.company});
        this.payload["stations"] = new StationContainer({ app: this.app, texture: textures.station});
        this.payload["rail_nodes"] = new RailNodeContainer({ app: this.app});
        this.payload["rail_edges"] = new RailEdgeContainer({ app: this.app});

        Object.keys(this.payload).forEach(key => {
            this.payload[key].setupDefaultValues();
            this.payload[key].setupUpdateCallback();
            this.payload[key].setupBeforeCallback();
            this.payload[key].setupAfterCallback();
            this.payload[key].begin();
        });
    }

    protected viewDebugInfo() {
        this.debugText.text = "FPS: " + this.app.ticker.FPS.toFixed(2)
                                + ", " + this.app.stage.children.length + " entities"
                                + ", debug=" + this.debugValue;
    }

    /**
     * 指定した id に対応するリソースを取得します
     * @param key リソース型
     * @param id id
     */
    get(key: string, id: string) {
        let container = this.payload[key];
        if (container !== undefined) {
            return container.getChild(id);
        }
        return undefined;
    }

    mergeAll(payload: GameMap) {
        config.zIndices.forEach(key => {
            if (this.payload[key] !== undefined) {
                this.payload[key].mergeChildren(payload[key], {coord: this.coord});
                if (this.payload[key].isChanged()) {
                    this.changed = true;
                }
            }
        });
        this.resolve();
    }

    resolve() {
        if (this.payload["rail_nodes"] !== undefined) {
            this.payload["rail_nodes"].forEachChild((rn : RailNode) => {
                rn.resolve(this.get("rail_nodes", rn.get("pid")))
            });
        }
        if (this.payload["rail_edges"] !== undefined) {
            this.payload["rail_edges"].forEachChild((re: RailEdge) => 
                re.resolve(
                    this.get("rail_nodes", re.get("from")),
                    this.get("rail_nodes", re.get("to")),
                    this.get("rail_edges", re.get("eid"))
                )
            );
        }
    }

    setCenter(x: number, y: number, force: boolean = false) {
        let short = Math.min(this.renderer.width, this.renderer.height);
        let long = Math.max(this.renderer.width, this.renderer.height);
        let shortRadius = Math.pow(2, this.coord.scale - 1 + Math.log2(short/long));
        let longRadius = Math.pow(2, this.coord.scale - 1);

        if (this.renderer.width < this.renderer.height) {
            // 縦長
            if (x - shortRadius < config.gamePos.min.x) {
                x = config.gamePos.min.x + shortRadius;
            }
            if (x + shortRadius > config.gamePos.max.x) {
                x = config.gamePos.max.x - shortRadius;
            }
            if (y - longRadius < config.gamePos.min.y) {
                y = config.gamePos.min.y + longRadius;
            }
            if (y + longRadius > config.gamePos.max.y) {
                y = config.gamePos.max.y - longRadius;
            }
            if (this.coord.scale > config.scale.max) { 
                y = 0;
            }
        }else {
            // 横長
            if (x - longRadius < config.gamePos.min.x) {
                x = config.gamePos.min.x + longRadius;
            }
            if (x + longRadius > config.gamePos.max.x) {
                x = config.gamePos.max.x - longRadius;
            }
            if (y - shortRadius < config.gamePos.min.y) {
                y = config.gamePos.min.y + shortRadius;
            }
            if (y + shortRadius > config.gamePos.max.y) {
                y = config.gamePos.max.y - shortRadius;
            }
            if (this.coord.scale > config.scale.max) { 
                x = 0;
            }
        }
        if (this.coord.cx == x && this.coord.cy == y) {
            return;
        }
        this.coord.cx = x;
        this.coord.cy = y;
        
        this.updateCoord(force);
    }

    setScale(v: number, force: boolean = false) {
        let old = this.coord.scale

        let short = Math.min(this.renderer.width, this.renderer.height);
        let long = Math.max(this.renderer.width, this.renderer.height);
        let maxScale = config.scale.max + Math.log2(long/short);

        if (v < config.scale.min) {
            v = config.scale.min;
        }
        if (v > maxScale) {
            v = maxScale;
        }
        this.coord.zoom = v < old ? 1 : v > old ? -1 : 0;
        if (this.coord.scale == v) {
            return;
        } 
        this.coord.scale = v;

        this.updateCoord(force);
    }

    protected updateCoord(force: boolean) {
        Object.keys(this.payload).forEach(key => {
            this.payload[key].merge("coord", this.coord);
            if (force) {
                this.payload[key].mergeAll(forceMove);
            }
            if (this.payload[key].isChanged()) {
                this.changed = true;
            }
        });
        this.world.merge("coord", this.coord);
    }

    isChanged() {
        return this.changed;
    }

    render() {
        Object.keys(this.payload).forEach(key => 
            this.payload[key].forEachChild((c) => c.beforeRender())
        );
        Object.keys(this.payload).forEach(key => this.payload[key].reset());
        this.changed = false;
    }

    unmount() {
        Object.keys(this.payload).reverse().forEach(key => this.payload[key].end());
        this.cursor.end();
        this.world.end();
    }
}
