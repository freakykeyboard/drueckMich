'use strict';
//ToDo hasOwnProperty
window.addEventListener("load", () => {

    setInterval(updateBookmarks, 1000 * 5);

    function updateBookmarks() {
        let xhr = new XMLHttpRequest();
        const url = "update";

        xhr.addEventListener("load", () => {
            let tbody;
            let response = JSON.parse(xhr.responseText);

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
                let temp, tds, clone, a, img, p;
                //get the tenmplate element
                temp = document.getElementsByTagName("template")[0];
                tds = temp.content.querySelectorAll("td");
                a = temp.content.querySelector("a");
                console.log(tds.length);

                for (let item in response.bookmarks) {
                    let bookmark = response.bookmarks[item];
                    a.setAttribute("href", bookmark.url);


                    a.textContent = bookmark.url;
                    tds[1].textContent = bookmark.shortReview;
                    tds[2].textContent = bookmark.title;
                    tds[3].textContent = bookmark.images;
                    img=tds[4].querySelector("img");
                    img.setAttribute("src","/gridGetIcon?fileName="+bookmark.icon);
                    img.content=bookmark.icon;
                    tds[5].textContent = bookmark.wvr_categories;


                    tds[6].textContent = bookmark.custom_categorie;
                    p = tds[7].querySelectorAll("p");
                    p[0].textContent = "Latitude " + bookmark.lat;
                    p[1].textContent = "Longitude " + bookmark.long;
                    clone = document.importNode(temp.content, true);
                    tbody.appendChild(clone)
                }

            }


        });
        xhr.open("GET", url, true);
        xhr.send();
    }


});

function titleClick(e) {
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