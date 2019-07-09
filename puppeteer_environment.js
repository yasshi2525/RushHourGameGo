const JsdomEnvironment = require("jest-environment-jsdom");
const fs = require("fs");
const path = require("path");
const puppeteer = require("puppeteer");
const os = require("os");

const DIR = path.join(os.tmpdir(), "jest_puppeteer_global_setup");

class PuppeteerEnvironment extends JsdomEnvironment {
  constructor(config) {
    super(config);
  }

  async setup() {
    await super.setup();
    // get the wsEndpoint
    const wsEndpoint = fs.readFileSync(path.join(DIR, "wsEndpoint"), "utf8");
    if (!wsEndpoint) {
      throw new Error("wsEndpoint not found");
    }

    // connect to puppeteer
    const browser = await puppeteer.connect({
      browserWSEndpoint: wsEndpoint,
    });
    
    this.global.__BROWSER__ = browser;
  }

  async teardown() {
    await super.teardown();
  }

  runScript(script) {
    return super.runScript(script);
  }
}

module.exports = PuppeteerEnvironment;