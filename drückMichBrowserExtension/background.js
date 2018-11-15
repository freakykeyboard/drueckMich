'use strict';
//ToDo change " to '
document.addEventListener('DOMContentLoaded',()=>{
   //browser action button handler
   chrome.browserAction.onClicked.addListener((tab)=>{
        chrome.tabs.query({
            active:true,
            currentWindow:true
        },(tabs)=>{
            let activeTab=tabs[0];
            chrome.tabs.sendMessage(activeTab.id,{
                "message":"extensionButtonWasPressed"
            })
        })
   });
   chrome.runtime.onMessage.addListener((request,sender,sendResponse)=>{
       if (request.message==="sendHref"){


           let url='http://localhost:4242/Url?url='+request.href;
           console.log(url);
           let xhr=new XMLHttpRequest();
           xhr.open('POST',url);

           xhr.setRequestHeader("Content-type", 'application/x-www-form-urlencoded');
           xhr.send();

           // Neuen Tab mit urlNeuerTab öffnen. Falls diese Url bereits in einem Tab geöffnet ist,
           // diesen Tab aktivieren und neu laden:
           let newTabUrl = "localhost:4242/drückMich";

           chrome.tabs.query({
               url: newTabUrl
           }, function (tabs) {
               if (tabs[0]) {
                   // es existiert ein Tab mit dieser urlNeuerTab:
                   // -> Tab aktivieren
                   // -> Tab neu laden
                   chrome.tabs.update(tabs[0].id, {
                       active: true
                   });
                   chrome.tabs.reload(tabs[0].id);
               } else {
                   // neuen Tab mit urlNeuerTab öffnen:
                   chrome.tabs.create({
                       url: newTabUrl
                   });
               }
           });
       }
   })
});