import * as PIXI from "pixi.js";
import { config, Coordinates, Point } from "../interfaces/gamemap";
import { Monitorable, MonitorContrainer } from "../interfaces/monitor";
import { ApplicationProperty } from "../interfaces/pixi";
import BaseContainer from "./container";
import BaseModel from "./base";

const defaultValues: {coord: Coordinates, [index: string]: any} = {
    coord:  {
        cx: config.gamePos.default.x, 
        cy: config.gamePos.default.y, 
        scale: config.scale.default,
        zoom: 0
    },
    forceMove: false,
    outMap: false,
    visible: true
};

export abstract class PIXIModel extends BaseModel implements Monitorable {
    protected app: PIXI.Application;
    protected parent: PIXI.Container;
    protected container: PIXI.Container;
    /**
     * smoothMove後、描画する座標(クライアント座標系)
     */
    destination: Point;
    /**
     * 描画する座標(クライアント座標系)
     */
    current: Point;
    /**
     * (x, y)が変化したとき、destination に移動するまでの残りフレーム数。
     */
    protected latency: number;

    protected smoothMoveFn: () => void;

    constructor(options: ApplicationProperty) {
        super();
        this.app = options.app;
        this.parent = options.container;
        this.container = new PIXI.Container();
        this.destination = {x: 0, y: 0};
        this.current = {x: 0, y: 0};
        this.latency = 0;
        this.smoothMoveFn = () => this.smoothMove();
    }

    setupDefaultValues() {
        super.setupDefaultValues();
        this.addDefaultValues(defaultValues);
    }

    setupBeforeCallback() {
        super.setupBeforeCallback();
        this.addBeforeCallback(() => {
            this.app.stage.addChild(this.container);
            this.app.ticker.add(this.smoothMoveFn);
            this.container.visible = this.props.visible;
        })
    }
    
    setupUpdateCallback() {
        super.setupUpdateCallback();
        this.addUpdateCallback("coord", () => this.updateDestination());
        this.addUpdateCallback("forceMove", (v) => {
            if (v) {
                this.moveDestination();
            }
        });
        this.addUpdateCallback("visible", (v) => {this.container.visible = v});
    }

    setupAfterCallback() {
        super.setupAfterCallback();
        this.addAfterCallback(() => {
            this.app.ticker.remove(this.smoothMoveFn);
            this.app.stage.removeChild(this.container);
        })
    }

    toView(x: number, y: number): Point {
        let center = {
            x: this.app.renderer.width / this.app.renderer.resolution / 2,
            y: this.app.renderer.height / this.app.renderer.resolution / 2,
        }
        let size = Math.max(this.app.renderer.width / this.app.renderer.resolution, this.app.renderer.height / this.app.renderer.resolution)
        let zoom = Math.pow(2, -this.props.coord.scale)

        return {
            x: (x - this.props.coord.cx) * size * zoom + center.x,
            y: (y - this.props.coord.cy) * size * zoom + center.y
        }
    }

    shouldEnd() {
        return this.props.outMap && this.current == this.destination;
    }

    protected calcDestination() {
        return this.toView(this.props.x, this.props.y);
    }

    updateDestination() {
        this.destination = this.calcDestination();
        this.latency = config.latency;
    }

    moveDestination() {
        this.current = this.destination;
        this.latency = 0;
        this.props.forceMove = false;
    }

    protected smoothMove() {   
        if (this.latency > 0) {
            let ratio = this.latency / config.latency;
            if (ratio < 0.5) {
                ratio = 1.0 - ratio;
            }
            this.current.x = this.current.x * ratio + this.destination.x * (1 - ratio);
            this.current.y = this.current.y * ratio + this.destination.y * (1 - ratio);
            this.latency--;
        } else {
            this.current = this.destination;
            this.latency = 0;
        }
        this.beforeRender();
    }
}

export abstract class PIXIContainer<T extends PIXIModel> extends BaseContainer<T> implements MonitorContrainer {
    protected app: PIXI.Application;

    constructor(
        options: ApplicationProperty,
        newInstance: { new (props: {[index:string]: {}}): T }, 
        newInstanceOptions: {[index:string]: {}}) {
        super(newInstance, newInstanceOptions);
        this.app = options.app;
        this.childOptions.app = this.app;
    }

    setupDefaultValues() {
        super.setupDefaultValues();
        this.addDefaultValues(defaultValues);
    }
}