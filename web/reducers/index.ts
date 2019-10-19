import * as Actions from "../actions";
import { ActionPayload, AsyncStatus } from "..//common/interfaces";
import { RushHourStatus } from "../state";

export default (state: RushHourStatus, action: {type: string, payload: ActionPayload}) => {
    switch (action.type) {
        case Actions.login.success.toString():
            return Object.assign({}, state, { isLoginSucceeded: true });
        case Actions.login.failure.toString():
            return Object.assign({}, state, { isLoginFailed: true });
        case Actions.resetLoginError.toString():
            return Object.assign({}, state, { isLoginFailed: false });
        case Actions.register.success.toString():
            return Object.assign({}, state, { isRegisterSucceeded: true });
        case Actions.register.failure.toString():
            return Object.assign({}, state, { isRegisterFailed: true });
        case Actions.settings.success.toString():
            return Object.assign({}, state, { settings: action.payload.results });
        case Actions.editSettings.request.toString():
            return Object.assign({}, state, { waitingFor: action.payload });
        case Actions.editSettings.success.toString():
            let my = Object.assign({}, state.my, action.payload.results.my)
            let settings = Object.assign({}, state.settings, { [action.payload.results.key]: action.payload.results.value });
            return Object.assign({}, state, { waitingFor: undefined, settings, my });
        case Actions.editSettings.failure.toString():
            return Object.assign({}, state, { waitingFor: undefined });
        case Actions.setMenu.success.toString():
            return Object.assign({}, state, { menu: action.payload });
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
        case Actions.gameStatus.request.toString():
        case Actions.inOperation.request.toString():
            var inOperation: AsyncStatus = Object.assign({}, state.inOperation, { waiting: true });
            return Object.assign({}, state, inOperation);
        case Actions.gameStatus.success.toString(): 
        case Actions.inOperation.success.toString():
            var inOperation: AsyncStatus = Object.assign({}, state.inOperation, { waiting: false, value: action.payload.results });
            return Object.assign({}, state, {inOperation});
        case Actions.gameStatus.failure.toString():
        case Actions.inOperation.failure.toString():
            var inOperation: AsyncStatus = Object.assign({}, state.inOperation, { waiting: false });
            return Object.assign({}, state, {inOperation});
        default:
            return state;
    }
};
