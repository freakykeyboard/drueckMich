'use strict';
// Dieser Code wird im Kontext einer fremden Seite (aktiver Tab) ausgeführt.
// Mit der eigenen Anwendung (background.js) kann nur über messages kommuniziert werden:
//
chrome.runtime.onMessage.addListener(
    function (request, sender, sendResponse) {
        if (request.message === "extensionButtonWasPressed") {
            console.log('butto was pressed');
            let hrefActiveTab = window.location.href;

            chrome.runtime.sendMessage({
                "message": "sendHref",
                "href": hrefActiveTab
            });
        }
    }
);