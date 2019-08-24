import * as PIXI from "pixi.js";
import { AnimatedSpriteModel } from "./sprite";
import { Monitorable } from "../interfaces/monitor";
import { MenuStatus } from "../../state";
import { PIXIProperty } from "../interfaces/pixi";
import { GraphicsAnimationGenerator } from "./animate";
import { PointModel } from "./point";
import { RailNode } from "./rail";
import Anchor from "./anchor";
import { Point } from "../interfaces/gamemap";

const graphicsOpts = {
    padding: 20,
    width: 1,
    alpha: 0.2,
    color: 0x607d8B,
    radius: 20
};

const defaultValues: {
    menu: MenuStatus,
    client: Point
} = {
    menu: MenuStatus.IDLE,
    client: {x: 0, y: 0}
};

export default class extends AnimatedSpriteModel implements Monitorable {
    selected: PointModel | undefined;
    anchor: Anchor;

    constructor(options: PIXIProperty & { offset: number, anchor: Anchor } ) {
        let graphics = new PIXI.Graphics();
        graphics.lineStyle(graphicsOpts.width, graphicsOpts.color);
        graphics.beginFill(graphicsOpts.color, graphicsOpts.alpha);
        graphics.drawCircle(
            graphicsOpts.padding + graphicsOpts.radius,
            graphicsOpts.padding + graphicsOpts.radius,
            graphicsOpts.radius 
        );
        graphics.endFill();

        let generator = new GraphicsAnimationGenerator(options.app, graphics);

        let rect = graphics.getBounds().clone();
        rect.x -= graphicsOpts.padding - 1;
        rect.y -= graphicsOpts.padding - 1;
        rect.width += graphicsOpts.padding * 2;
        rect.height += graphicsOpts.padding * 2;

        let animation = generator.record(rect);

        super({ animation, ...options });
        this.anchor = options.anchor;
    }
    
    setupDefaultValues() {
        super.setupDefaultValues();
        this.addDefaultValues(defaultValues);
    }

    setInitialValues(props: {[index: string]: any}) {
        super.setInitialValues(props);
        this.props.client = undefined;
        this.props.pos = undefined;
        this.updateDestination();
        this.moveDestination();
    }

    setupUpdateCallback() {
        super.setupUpdateCallback();
        this.addUpdateCallback("client", (v) => {
            this.merge("pos", this.toServer(v, 2))
            this.selectObject();
            this.moveDestination();
        });
        this.addUpdateCallback("coord", () => {
            this.merge("pos", this.toServer(this.props.client, 2))
            this.selectObject();
            this.updateDestination();
        });
    }

    protected calcDestination() {
        return (this.selected === undefined) 
        ? this.toView(this.toServer(this.props.client, 2))
        : this.toView(this.selected.get("pos"));
    }


    updateDisplayInfo() {
        if (!this.isVisible()) {
            this.sprite.visible = false;
            return;
        }
        if (!this.followPointModel(this.selected, 3)) {
            super.updateDisplayInfo();
        }
    }

    selectObject(except: PointModel | undefined = undefined) {
        if (this.props.pos === undefined) {
            this.unlinkSelected();
            return;
        }
        var selected;
        switch(this.props.menu) {
            case MenuStatus.SEEK_DEPARTURE:
            case MenuStatus.EXTEND_RAIL:
                selected = this.model.gamemap.getOnChunk("rail_nodes", this.props.pos, 1);
                break;
        }
        if (selected instanceof PointModel && selected !== this.anchor.object && selected !== except) {
            this.selected = selected;
            selected.refferedCursor = this;
            this.updateDestination();
        } else {
            this.unlinkSelected();
        }
    }

    unlinkSelected() {
        if (this.selected !== undefined) {
            this.selected = undefined;
            this.updateDestination();
        }
    }

    genAnchorStatus() {
        if (this.selected === undefined) {
            return undefined;
        } else {
            let res = { pos: this.selected.get("pos"), type: "", cid: this.selected.get("cid") };
            if (this.selected instanceof RailNode) {
                res.type = "rail_nodes"
                return res
            } 
            return undefined;
        }
    }

    protected isVisible() {
        return this.props.menu === MenuStatus.SEEK_DEPARTURE || this.props.menu === MenuStatus.EXTEND_RAIL;
    }
}