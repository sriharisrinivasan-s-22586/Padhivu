import {fetchPost} from "../util/fetch";

const getLANConfig = () => {
    if (!window.siyuan.config.sync.lan) {
        window.siyuan.config.sync.lan = {
            endpoint: "",
            basePath: window.siyuan.config.system.dataDir || "",
            authToken: "",
            deviceID: window.siyuan.config.system.id || "",
            deviceName: window.siyuan.config.system.name || "",
            serve: false,
            timeout: 15,
            concurrentReqs: 4,
        };
    }
    return window.siyuan.config.sync.lan;
};

export const renderLANProvider = () => {
    const lan = getLANConfig();
    const hostMode = !!lan.serve;
    return `<div class="b3-label b3-label--inner">
    Built-in LAN sync for your own devices. One device can act as the host and other devices connect to it over your local network.
</div>
<div class="b3-label b3-label--inner fn__flex">
    <div class="fn__flex-center fn__size200">Host Endpoint</div>
    <div class="fn__space"></div>
    <input id="lanEndpoint" class="b3-text-field fn__block" placeholder="http://192.168.1.10:6806" value="${lan.endpoint || ""}"${hostMode ? " disabled" : ""}>
</div>
<div class="b3-label b3-label--inner fn__flex">
    <div class="fn__flex-center fn__size200">Host Mode</div>
    <div class="fn__space"></div>
    <input id="lanServe" class="b3-switch fn__flex-center" type="checkbox"${lan.serve ? " checked" : ""}>
</div>
<div class="b3-label b3-label--inner fn__flex">
    <div class="fn__flex-center fn__size200">Host Notes Path</div>
    <div class="fn__space"></div>
    <input id="lanBasePath" class="b3-text-field fn__block" value="${window.siyuan.config.system.dataDir || lan.basePath || ""}" disabled>
    <div class="b3-label__text">Uses the current workspace data directory automatically while Host Mode is enabled.</div>
</div>
<div class="b3-label b3-label--inner fn__flex">
    <div class="fn__flex-center fn__size200">Shared Token</div>
    <div class="fn__space"></div>
    <div class="b3-form__icona fn__block">
        <input id="lanAuthToken" type="password" class="b3-text-field b3-form__icona-input" value="${lan.authToken || ""}">
        <svg class="b3-form__icona-icon" data-action="togglePassword"><use xlink:href="#iconEye"></use></svg>
    </div>
</div>
<div class="b3-label b3-label--inner fn__flex">
    <div class="fn__flex-center fn__size200">Device Name</div>
    <div class="fn__space"></div>
    <input id="lanDeviceName" class="b3-text-field fn__block" value="${lan.deviceName || window.siyuan.config.system.name || ""}">
</div>
<div class="b3-label b3-label--inner fn__flex">
    <div class="fn__flex-center fn__size200">Device ID</div>
    <div class="fn__space"></div>
    <input id="lanDeviceID" class="b3-text-field fn__block" value="${lan.deviceID || window.siyuan.config.system.id || ""}">
</div>
<div class="b3-label b3-label--inner fn__flex">
    <div class="fn__flex-center fn__size200">Timeout (s)</div>
    <div class="fn__space"></div>
    <input id="lanTimeout" class="b3-text-field fn__block" type="number" min="7" max="300" value="${lan.timeout || 15}">
</div>
<div class="b3-label b3-label--inner fn__flex">
    <div class="fn__flex-center fn__size200">Concurrent Reqs</div>
    <div class="fn__space"></div>
    <input id="lanConcurrentReqs" class="b3-text-field fn__block" type="number" min="1" max="16" value="${lan.concurrentReqs || 4}">
</div>
<div class="b3-label b3-label--inner">
    <div class="fn__flex">
        <div class="fn__flex-1">Connected Devices</div>
        <button class="b3-button b3-button--outline fn__size200" data-action="refreshLanDevices">
            <svg><use xlink:href="#iconRefresh"></use></svg>Refresh
        </button>
    </div>
    <div id="lanConnectedDevices" class="fn__hr--b"></div>
</div>`;
};

export const saveLANProvider = (providerPanelElement: Element, cb?: (lan: typeof window.siyuan.config.sync.lan) => void) => {
    getLANConfig();
    let timeout = parseInt((providerPanelElement.querySelector("#lanTimeout") as HTMLInputElement).value, 10);
    if (7 > timeout) {
        timeout = 7;
    }
    if (300 < timeout) {
        timeout = 300;
    }
    let concurrentReqs = parseInt((providerPanelElement.querySelector("#lanConcurrentReqs") as HTMLInputElement).value, 10);
    if (1 > concurrentReqs) {
        concurrentReqs = 1;
    }
    if (16 < concurrentReqs) {
        concurrentReqs = 16;
    }
    const endpointElement = providerPanelElement.querySelector("#lanEndpoint") as HTMLInputElement;
    let endpoint = endpointElement.value.trim();
    if (endpoint && !endpoint.startsWith("http")) {
        endpoint = "http://" + endpoint;
    }
    endpointElement.value = endpoint;
    const lan = {
        endpoint,
        basePath: window.siyuan.config.system.dataDir || "",
        authToken: (providerPanelElement.querySelector("#lanAuthToken") as HTMLInputElement).value.trim(),
        deviceName: (providerPanelElement.querySelector("#lanDeviceName") as HTMLInputElement).value.trim(),
        deviceID: (providerPanelElement.querySelector("#lanDeviceID") as HTMLInputElement).value.trim(),
        serve: (providerPanelElement.querySelector("#lanServe") as HTMLInputElement).checked,
        timeout,
        concurrentReqs,
    };
    fetchPost("/api/sync/setSyncProviderLAN", {lan}, (response) => {
        window.siyuan.config.sync.lan = response.data.lan;
        cb?.(response.data.lan);
    });
};

export const toggleLANModeFields = (providerPanelElement: Element) => {
    const hostMode = (providerPanelElement.querySelector("#lanServe") as HTMLInputElement)?.checked;
    const endpoint = providerPanelElement.querySelector("#lanEndpoint") as HTMLInputElement;
    if (endpoint) {
        endpoint.disabled = !!hostMode;
    }
};

export const renderLANDevices = (container: Element, devices?: Array<{
    id: string;
    name: string;
    host: string;
    os: string;
    role: string;
    connected: boolean;
    lastSeen: number;
}>) => {
    if (!devices || devices.length === 0) {
        container.innerHTML = `<div class="ft__on-surface">No LAN sync devices have connected yet.</div>`;
        return;
    }
    container.innerHTML = `<div class="b3-list">${devices.map((device) => {
        return `<div class="b3-list-item b3-list-item--two" style="cursor:auto">
    <div class="b3-list-item__first">
        <span class="b3-list-item__text">${device.name || device.id}</span>
        <span class="b3-list-item__meta">${device.role}</span>
    </div>
    <div class="b3-list-item__meta">${device.connected ? "Online" : "Recent"}</div>
    <div class="b3-list-item__text ft__on-surface">${device.host || device.os || ""}</div>
</div>`;
    }).join("")}</div>`;
};

export const refreshLANDevices = (providerPanelElement: Element) => {
    const container = providerPanelElement.querySelector("#lanConnectedDevices");
    if (!container) {
        return;
    }
    fetchPost("/api/sync/getSyncInfo", {}, (response) => {
        renderLANDevices(container, response.data.lanDevices);
    });
};
