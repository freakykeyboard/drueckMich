'use strict';
//ToDo hasOwnProperty
window.addEventListener("load", () => {

    setInterval(updateBookmarks, 1000 * 5);

    function updateBookmarks() {
        let xhr = new XMLHttpRequest();
        const url = "update";

        xhr.addEventListener("load", () => {

            updateTable(JSON.parse(xhr.responseText));



        });
        xhr.open("GET", url, true);
        xhr.send();
    }


});

function titleClick() {
    let url = "setSortProperties";
    let formData = new FormData();
    let xhr = new XMLHttpRequest();
    xhr.addEventListener("load", () => {
        console.log(xhr.responseText)
    });
    xhr.open("POST", url);

    formData.append('orderBy', "0");
    xhr.send(formData);
    console.log('clicked')
}
function updateTable(response) {
    console.log(response);
    let tbody;
    let table = document.getElementById("bookmarkTable");
    try {
        for (let child in table.childNodes) {
            if (table.childNodes[child].nodeName === "TBODY") {
                tbody = table.childNodes[child];
                while (table.childNodes[child].firstChild.nextSibling) {
                    table.childNodes[child].removeChild(table.childNodes[child].firstChild)
                }
                table.childNodes[child].removeChild(table.childNodes[child].firstChild)
            }
        }
    } catch (e) {

    }

    if (response.bookmarks) {
        console.log('response.bookmarks',response.bookmarks);
        let temp, td, clone, addImg,rmvImg, ulWvr,img, liWvr,ulCustom,liCustom, a, p1, p2, d;
        //get the tenmplate element
        temp = document.querySelector('#bookmarkRow');
        td = temp.content.querySelectorAll("td");
        a=temp.content.querySelector("a");

        addImg=temp.content.querySelector('#add');
        rmvImg=temp.content.querySelector('#remove');
        ulWvr=temp.content.querySelector('#wvrUl');
        liWvr=temp.content.querySelector("#wvrLi");
        ulCustom=temp.content.querySelector('#customUl');
        liCustom=temp.content.querySelector("#customLi");
        p1=temp.content.querySelectorAll("p")[0];
        p2=temp.content.querySelectorAll("p")[1];
        for (let i = 0; i < response.bookmarks.length; i++) {
            let bookmark=response.bookmarks[i];
            console.log(a);
            a.setAttribute("src",bookmark.url);
            a.textContent=bookmark.url;
            td[0].appendChild(a);
            td[0].textContent=bookmark.url;

            td[1].textContent=bookmark.shortReview;
            td[2].textContent=bookmark.title;

            td[3].textContent=bookmark.images;
            img=temp.content.querySelector('#icon');
            console.log(img);
            img.setAttribute("src","/gridGetIcon?fileName="+bookmark.icon);
            img.textContent="test";
            td[4].appendChild(img);
            for (let j=0;j<bookmark.wvr_categories;j++){
                liWvr.textContent=bookmark.wvr_categories[j];
                ulWvr.appendChild(liWvr);
            }
            td[5].appendChild(ulWvr);

            addImg=temp.content.querySelector("img");
            addImg.setAttribute("src","plusIcon.png");
            addImg.setAttribute("onclick","addCategoryToBookmark(event)");
            td[6].appendChild(addImg);


            for (let j=0;j<bookmark.custom_categories;j++){
                liCustom.textContent=bookmark.custom_categories[j];
                ulCustom.appendChild(liCustom);
            }
            td[6].appendChild(ulCustom);

            p1.textContent=bookmark.latitude;
            p2.textContent=bookmark.longitude;

            td[7].appendChild(p1);
            td[7].appendChild(p2);
            clone=document.importNode(temp.content,true)
            tbody.appendChild(clone);
        }

    }
}

