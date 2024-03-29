'use strict';

document.addEventListener('DOMContentLoaded',()=>{
   //browser action button handler
   chrome.browserAction.onClicked.addListener(()=>{
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

           let encodedeUrl=encodeURIComponent(request.href);
           let url='http://localhost:4242/Url?url='+encodedeUrl;
           let xhr=new XMLHttpRequest();

           xhr.open('GET',url);
           xhr.addEventListener("load",()=>{

              console.log(xhr.responseText)
           });
           xhr.setRequestHeader("Content-type", 'application/x-www-form-urlencoded');
           xhr.send();

           // Neuen Tab mit urlNeuerTab öffnen. Falls diese Url bereits in einem Tab geöffnet ist,
           // diesen Tab aktivieren und neu laden:
           var urlNeuerTab = "http://localhost:4242/drueckMich";

           chrome.tabs.query({
               url: urlNeuerTab
           }, function (tabs) {

               if (tabs[0]) {
                   console.log('tabs[0]',tabs[0])
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
                       url: urlNeuerTab
                   });
               }
           });

       }
   })
});
