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
function updateTable(response){
    console.log(response);
    let tbody;
    let table = document.getElementById("bookmarkTable");
    for (let child in table.childNodes) {
        if (table.childNodes[child].nodeName === "TBODY") {
            tbody = table.childNodes[child];
            while (table.childNodes[child].firstChild.nextSibling) {
                table.childNodes[child].removeChild(table.childNodes[child].firstChild)
            }
            table.childNodes[child].removeChild(table.childNodes[child].firstChild)
        }
    }
    if (response.bookmarks) {
        let temp, tds, clone,item, a, img, p,ul,li;
        //get the tenmplate element
        temp = document.getElementsByTagName("template")[0];
        tds = temp.content.querySelectorAll("td");
        a = temp.content.querySelector("a");

        console.log(tds.length);


        for (let i in response.bookmarks) {

            let bookmark = response.bookmarks[i];
            a.setAttribute("href", bookmark.url);
            a.textContent = bookmark.url;
            tds[1].textContent = bookmark.shortReview;
            tds[2].textContent = bookmark.title;
            tds[3].textContent = bookmark.images;
            img=tds[4].querySelector("img");
            img.setAttribute("src","/gridGetIcon?fileName="+bookmark.icon);
            img.content=bookmark.icon;
            tds[5].textContent = bookmark.wvr_categories;
            img=tds[6].querySelector("img");
            img.setAttribute("id",bookmark.url);
          p=tds[6].querySelector("p");
          p.textContent=bookmark.custom_categories;
            p = tds[7].querySelectorAll("p");
            p[0].textContent = "Latitude " + bookmark.lat;
            p[1].textContent = "Longitude " + bookmark.long;
            clone = document.importNode(temp.content, true);
            tbody.appendChild(clone)
        }

    }
}
