'use strict';

window.addEventListener("load",()=>{

   setInterval(updateBookmarks,1000*5);
    function updateBookmarks () {
        let xhr=new XMLHttpRequest();
        const url="update";

        xhr.addEventListener("load",()=>{
            let tbody;
            let response=JSON.parse(xhr.responseText);
            console.log('response',response)
            let table=document.getElementById("bookmarkTable");
            for(let child in table.childNodes){
                if(table.childNodes[child].nodeName==="TBODY"){
                    tbody=table.childNodes[child];
                    while(table.childNodes[child].firstChild.nextSibling){
                        table.childNodes[child].removeChild(table.childNodes[child].firstChild)
                    }
                    table.childNodes[child].removeChild(table.childNodes[child].firstChild)
                }
            }
            if (response.bookmarks) {
                console.log('response.bookmarks before sorting',response.bookmarks);
                    response.bookmarks.sort(function(a,b){
                   let x=a.title.toLowerCase();
                   let y=b.title.toLowerCase();
                   if (x<y)return -1;
                   if (x>y)return 1;
                   return 0;
                });
                console.log('response.bookmarks before sorting',response.bookmarks);
                for (let item in response.bookmarks){
                    let bookmark=response.bookmarks[item];

                    let tr=document.createElement("tr");
                    for (let i=0;i<8;i++){
                        let td=document.createElement("td");
                        switch (i) {
                            case 0:
                                let a=document.createElement("a");
                                a.setAttribute("target","_blank")
                                a.setAttribute("href",bookmark.url);
                                a.innerText=bookmark.url;
                                td.appendChild(a)

                                break;

                            case 1:
                                td.innerText=bookmark.shortReview;
                                break;
                            case 2:
                                td.innerText=bookmark.title;
                                break;
                            case 3:
                                td.innerText=bookmark.images;
                                break;

                            case 4:
                                let img=document.createElement("img");
                                img.setAttribute("src","gridGetIcon?fileName="+bookmark.icon);
                                td.appendChild(img)
                                break;

                            case 5:
                                td.innerText=bookmark.wvr_categories;
                                break;

                            case 6:
                                td.innerText=bookmark.custom_categories;
                                break;

                            case 7:
                                let p=document.createElement("P");


                                let text=document.createTextNode("Latitude");
                                p.appendChild(text);
                                td.appendChild(p);
                                text=document.createTextNode(bookmark.lat);
                                td.appendChild(text)
                                let br=document.createElement("br");
                                td.appendChild(br)
                                p=document.createElement("p");
                                text=document.createTextNode("Longitude");
                                p.appendChild(text)
                                td.appendChild(p)
                                text=document.createTextNode(bookmark.long);
                                td.appendChild(text)
                                break;



                            default:
                        }
                                tr.appendChild(td)
                        tbody.appendChild(tr)
                        table.appendChild(tbody)



                    }
                }
            }



        });
        xhr.open("GET",url,true);
        xhr.send();
    }


});

function titleClick(e){
    let url="setSortProperties";
    let formData=new FormData();
    let xhr=new XMLHttpRequest();
    xhr.addEventListener("load",()=>{
       console.log(xhr.responseText)
    });
    xhr.open("POST",url);

    formData.append('orderBy','title');
    xhr.send(formData);
console.log('clicked')
}