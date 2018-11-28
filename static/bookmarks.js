'use strict';

window.addEventListener("load",()=>{
    let names=["URL","Kurzbeschreibung","TitelText","Images","Icon/Logo","Kategorien","GPS"];
   //setInterval(updateBookmarks,1000);
    function updateBookmarks () {
        let xhr=new XMLHttpRequest();
        const url="update";

        xhr.addEventListener("load",()=>{
            let items=JSON.parse(xhr.responseText);
            console.log(items.bookmarks);
            let table=create("table");
            let tr=create("tr");
            let div=document.getElementById("tableDiv");
            div.appendChild(table);
            for (let i=0;i<6;i++){
                let th=create("th");
                th.innerText=names[i];
                tr.appendChild(th);
                table.appendChild(tr);
            }
            if (items.bookmarks){
                for (let item in items.bookmarks){
                    let bookmark=JSON.parse(items.bookmarks[item]);
                    let tr=create("tr");
                    tr.innerText=bookmark[item];


                }
            }

        });
        xhr.open("GET",url,true);
        xhr.send();
    }
    function create(name){
        return document.createElement(name);
    }
});

