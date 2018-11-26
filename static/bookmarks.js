'use strict';

window.addEventListener("load",()=>{
   setInterval(updateBookmarks,1000)
});
function updateBookmarks () {
let xhr=new XMLHttpRequest();
const url="update";

xhr.addEventListener("load",()=>{
    document.body.innerHTML=xhr.responseText;
});
xhr.open("GET",url,true);
    xhr.send();
}
