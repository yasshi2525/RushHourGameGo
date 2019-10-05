import * as PIXI from "pixi.js";
import GameModel from "./models";
import { config } from "./interfaces/gamemap";

export default class {
    app: PIXI.Application;
    protected images = ["station", "train"];
    sheets = ["cursor", "anchor", "residence", "company", "rail_node", "rail_edge"];
    model: GameModel;

    constructor(myid: number) {
        this.app = new PIXI.Application({
            width: window.innerWidth,
            height: window.innerHeight,
            backgroundColor: config.background,
            autoStart: true,
            antialias: true,
            resolution: window.devicePixelRatio,
            autoDensity: true
        });
        this.app.stage.sortableChildren = true;
        this.model = new GameModel({
            app: this.app, 
            cx: config.gamePos.default.x, 
            cy: config.gamePos.default.y, 
            scale: config.scale.default,
            zoom: 0,
            myid
        });

        this.images.forEach(key => this.app.loader.add(key, `public/img/${key}.png`));
        
        this.sheets.forEach(key => {
            this.app.loader.add(key, `public/spritesheet/${key}@${Math.floor(this.model.renderer.resolution)}x.json`);
        });
    }

    initModel() {
        this.model.init();
    }
}