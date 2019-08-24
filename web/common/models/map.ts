import { Monitorable } from "../interfaces/monitor";
import { GameMap, Identifiable } from "../../state";
import GroupModel from "./group";
import { ResidenceContainer, CompanyContainer } from "./background";
import { StationContainer } from "./station";
import { RailNodeContainer, RailEdgeContainer } from "./rail";
import { ZIndex } from "../interfaces/pixi";
import { PlayerContainer } from "./player";
import { ResolveError } from "../interfaces/gamemap";

export default class extends GroupModel {
    init() {
        let textures = this.model.app.loader.resources;
        let base = { model: this.model, app: this.model.app };
        this.containers["players"] = new PlayerContainer(base);
        this.containers["residences"] = new ResidenceContainer({ ...base, zIndex: ZIndex.RESIDENCE, texture: textures["residence"].texture});
        this.containers["companies"] = new CompanyContainer({ ...base, zIndex: ZIndex.COMPANY, texture: textures["company"].texture});
        this.containers["stations"] = new StationContainer({ ...base, zIndex: ZIndex.STATION, texture: textures["station"].texture});
        this.containers["rail_nodes"] = new RailNodeContainer({ ...base, zIndex: ZIndex.RAIL_NODE });
        this.containers["rail_edges"] = new RailEdgeContainer({  ...base, zIndex: ZIndex.RAIL_EDGE });
    
        super.init();
    }

    mergeChild(key: string, props: {id: string}): undefined | Monitorable {
        if (this.containers[key] === undefined) {
            return undefined;
        } 
        return this.containers[key].mergeChild(props);
    }

    mergeChildren(key: string, props: Identifiable[], opts: {[index: string]: any} = {}) {
        if (this.containers[key] !== undefined) {
            this.containers[key].mergeChildren(props, opts);
            if (this.containers[key].isChanged()) {
                this.changed = true;
            }
        }
    }

    mergeAll(payload: GameMap) {
        Object.keys(this.containers).forEach(key => {
            this.mergeChildren(key, payload[key], { coord: this.model.coord })
        });
        let error = this.resolve();
        this.model.controllers.updateAnchor();
        return error;
    }

    resolve() {
        let error: ResolveError = {};
        this.forEach(v => v.resolve(error));
        return error;
    }
}