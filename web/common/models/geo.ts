import * as PIXI from "pixi.js";
import { Monitorable } from "../interfaces/monitor";
import { ApplicationProperty } from "../interfaces/pixi";
import BaseModel from "./base";
import { config, Coordinates } from "../interfaces/gamemap";

const pixiDefaultValues: Coordinates & {[index: string]: any} = {
    cx: config.gamePos.default.x, 
    cy: config.gamePos.default.y, 
    scale: config.scale.default
};

export abstract class PIXIModel extends BaseModel implements Monitorable {
    protected app: PIXI.Application;
    protected container: PIXI.Container;

    constructor(options: ApplicationProperty) {
        super();
        this.app = options.app;
        this.container = new PIXI.Container();
    }

    setupDefaultValues() {
        super.setupDefaultValues();
        this.addDefaultValues(pixiDefaultValues);
    }

    setupBeforeCallback() {
        super.setupBeforeCallback();
        this.addBeforeCallback(() => {
            this.app.stage.addChild(this.container);
        })
    }

    setupAfterCallback() {
        super.setupAfterCallback();
        this.addAfterCallback(() => {
            this.app.stage.removeChild(this.container);
        })
    }

    protected toView(x: number, y: number): {x: number, y:number} {
        let center = {
            x: this.app.renderer.width / this.app.renderer.resolution / 2,
            y: this.app.renderer.height / this.app.renderer.resolution / 2,
        }
        let size = Math.max(this.app.renderer.width / this.app.renderer.resolution, this.app.renderer.height / this.app.renderer.resolution)
        let zoom = Math.pow(2, -this.props.scale)

        return {
            x: (x - this.props.cx) * size * zoom + center.x,
            y: (y - this.props.cy) * size * zoom + center.y
        }
    }

    /**
     * scale + 1 の範囲をキャッシュ保持領域としたとき、それを外れたかどうか判定する
     * @param x サーバ座標系x座標
     * @param y サーバ座標系y座標
     */
    protected isOut(x: number, y: number) {
        let zoom = Math.pow(2, this.props.scale);
        return Math.abs(x - this.props.cx) > zoom || Math.abs(y - this.props.cy) > zoom;
    }

    shouldEnd() {
        return this.isOut(this.props.x, this.props.y);
    }
}

const pointDefaultValues: {x: number, y:number} = {x: 0, y: 0};

export abstract class PointModel extends PIXIModel implements Monitorable {
    protected vx: number;
    protected vy: number;

    constructor(options: ApplicationProperty) {
        super(options);
        this.vx = 0;
        this.vy = 0;
    }

    setupDefaultValues() {
        super.setupDefaultValues();
        this.addDefaultValues(pointDefaultValues);
    }

    beforeRender() {
        super.beforeRender();
        let v = this.toView(this.props.x, this.props.y);
        this.vx = v.x;
        this.vy = v.y;
    }
}
