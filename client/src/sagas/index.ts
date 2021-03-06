import { takeEvery, takeLatest, put, call } from "redux-saga/effects";
import * as Action from "actions";
import { generatePIXI } from "./model";
import { generateMap } from "./map";
import { generateDepart, generateExtend, generateConnect } from "./rail";
import {
  generatePlayers,
  generateLogin,
  generateRegister,
  generateSettings,
  generateEditSettings,
  generatePlayersPlain,
  generateSignOut
} from "./player";
import { generateSetMenu } from "./menu";
import { generateDestroy } from "./destroy";
import {
  generateStatus,
  generateInOperation,
  generatePurgeUserData
} from "./admin";

export function* generateRequest(
  request: any,
  args: { [index: string]: any; payload: any },
  callbacks: { success: any; failure: any }
) {
  try {
    const response = yield call(request, args.payload);
    return yield put(callbacks.success(response));
  } catch (e) {
    return yield put(callbacks.failure(e));
  }
}

async function isResponseOK(response: Response) {
  if (!response.ok) {
    if (response.status === 401) {
      localStorage.removeItem("jwt");
      location.href = "/";
    }
    throw Error(response.statusText);
  }
  return response;
}

async function validateResponse(rawRes: Response) {
  let res = await isResponseOK(rawRes);
  return await res.json();
}

export enum Method {
  GET = "GET",
  PUT = "PUT",
  POST = "POST",
  DELETE = "DELETE"
}

export async function http(
  url: string,
  method: Method = Method.GET,
  params: { [index: string]: any } = {}
) {
  let headers = new Headers();
  let jwt = localStorage.getItem("jwt");
  if (jwt !== null) {
    headers.set("Authorization", `Bearer ${jwt}`);
  }
  if (method !== Method.GET) {
    headers.set("Content-type", "application/json");
  }

  let rawRes =
    method === Method.GET
      ? await fetch(url, { headers })
      : await fetch(url, {
          method,
          headers,
          body: JSON.stringify(params, (key, value) => {
            return key == "model" ? undefined : value;
          })
        });
  return await validateResponse(rawRes);
}

/**
 * 非同期処理呼び出す ActionType を指定する。
 * ここで定義した ActionTypeをキャッチした際、個々のtsで定義した非同期メソッドが呼び出される
 */
export function* rushHourSaga() {
  yield takeLatest(Action.initPIXI.request, generatePIXI);
  yield takeLatest(Action.fetchMap.request, generateMap);
  yield takeLatest(Action.login.request, generateLogin);
  yield takeLatest(Action.signout.request, generateSignOut);
  yield takeLatest(Action.register.request, generateRegister);
  yield takeLatest(Action.settings.request, generateSettings);
  yield takeLatest(Action.editSettings.request, generateEditSettings);
  yield takeLatest(Action.players.request, generatePlayers);
  yield takeLatest(Action.playersPlain.request, generatePlayersPlain);
  yield takeLatest(Action.depart.request, generateDepart);
  yield takeLatest(Action.extend.request, generateExtend);
  yield takeLatest(Action.connect.request, generateConnect);
  yield takeLatest(Action.destroy.request, generateDestroy);
  yield takeLatest(Action.gameStatus.request, generateStatus);
  yield takeLatest(Action.inOperation.request, generateInOperation);
  yield takeLatest(Action.purgeUserData.request, generatePurgeUserData);
  yield takeEvery(Action.setMenu.request, generateSetMenu);
}
