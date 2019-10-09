import * as Actions from "../actions";
import { ActionPayload } from "..//common/interfaces";
import { RushHourStatus } from "../state";

export default (state: RushHourStatus, action: {type: string, payload: ActionPayload}) => {
    switch (action.type) {
        case Actions.setMenu.success.toString():
            return Object.assign({}, state, { menu: action.payload })
        case Actions.fetchMap.success.toString():
            return Object.assign({}, state, { 
                timestamp: action.payload.timestamp,
                isPlayerFetched: !action.payload.hasUnresolvedOwner,
                isFetchRequired: false
            });
        case Actions.destroy.success.toString(): {
            return Object.assign({}, state, {
                timestamp: action.payload.timestamp,
                isFetchRequired: true
            });
        }
        default:
            return state;
    }
};
