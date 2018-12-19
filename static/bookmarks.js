'use strict';
window.addEventListener("load", () => {

    setInterval(update, 1000 * 5);
    //zyklische r Ajax-Request für das Update
    function update() {
        let xhr = new XMLHttpRequest();
        const url = "update";
        xhr.addEventListener("load", () => {
            updateContent(JSON.parse(xhr.responseText));
        });
        xhr.open("POST", url, true);
        //immer wenn Koordinaten eingeben FormData senden
        let form = document.getElementById('geospatial');
        let geoFormData = new FormData(form);
        if (geoFormData) {
            xhr.send(geoFormData);
        } else {
            xhr.send();
        }
    }
});
//ajax-request für das Hochladen exportierter Lesezeichen
function uploadFile(e) {
    e.preventDefault();
    let form = document.getElementById('uploadBookmark');
    let url = 'upload';
    let xhr = new XMLHttpRequest();
    xhr.addEventListener('load', () => {

    });
    xhr.open('POST', url);
    xhr.send(new FormData(form));
}
//ajax um zum Server dass nach Title sortiert werden soll
function sortAfterTitle() {
    let url = "setSortProperties";
    let formData = new FormData();
    let xhr = new XMLHttpRequest();
    xhr.addEventListener("load", () => {
        updateContent(JSON.parse(xhr.responseText))
    });
    xhr.open("POST", url);

    formData.append('orderBy', "0");
    xhr.send(formData);

}
//ajax um zum Server dass nachdem Erstelldatum sortiert werden soll
function sortAfterTime() {
    let url = "setSortProperties";
    let formData = new FormData();
    let xhr = new XMLHttpRequest();

    xhr.addEventListener("load", () => {

        updateContent(JSON.parse(xhr.responseText))
    });
    xhr.open("POST", url);

    formData.append('orderBy', "1");
    xhr.send(formData);
}
//geopsatial umkreissuche ajax-request
function geospatial(e) {
    e.preventDefault();
    let form = document.getElementById('geospatial');
    let formData = new FormData(form);

    let url = 'geospatial';
    let xhr = new XMLHttpRequest();
    xhr.addEventListener('load', () => {
        updateContent(JSON.parse(xhr.responseText))
    });
    xhr.open('post', url);
    xhr.send(formData);
}

function updateContent(data) {

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
    if (data.hasOwnProperty('available_categories')) {
        let select = document.getElementById('filterBookmarks');
        let index = select.selectedIndex;
        while (select.firstChild.nextSibling) {
            select.removeChild(select.firstChild);
        }
        select.removeChild(select.firstChild);
        let temp = document.getElementById('selectCategory').content.cloneNode(true);
        temp.querySelector('.category').innerText = "nach Kategorien filtern";

        select.appendChild(temp);

        for (let j in data.available_categories) {
            let category = data.available_categories[j];
            let temp = document.getElementById('selectCategory').content.cloneNode(true);
            temp.querySelector('.category').innerText = category;

            select.appendChild(temp);
        }

        select.selectedIndex = index;
    }
    if (data.bookmarks) {
        for (let i = 0; i < data.bookmarks.length; i++) {
            let bookmark = data.bookmarks[i];

            let temp = document.getElementById('bookmarkRow').content.cloneNode(true);

            let a = document.createElement('a');
            a.setAttribute('target', '_blank');
            a.innerText = bookmark.title;
            a.setAttribute('href', bookmark.url);
            temp.querySelector('.url').appendChild(a);
            temp.querySelector('.shortReview').innerText = bookmark.shortReview;
            console.log('bookmark.title', bookmark.title);
            temp.querySelector('.title').innerText = bookmark.title;

            for (let j in bookmark.wvr_categories) {
                let li = document.createElement('li');
                li.innerText = bookmark.wvr_categories[j];
                temp.querySelector('.wvrCategories').appendChild(li);
            }
            for (let j in bookmark.custom_categories) {
                let li = document.createElement('li');
                li.innerText = bookmark.custom_categories[j];
                temp.querySelector('.customCategories').appendChild(li);
            }
            temp.querySelector('.imgIcon').setAttribute('src', "/gridGetIcon?fileName=" + bookmark.icon);

            temp.querySelector('.lat').innerText += bookmark.lat;
            temp.querySelector('.lon').innerText += bookmark.long;
            temp.querySelector('.addCategory').setAttribute('id', bookmark.url);
            temp.querySelector('.customCategories').setAttribute("id", bookmark.url);
            temp.querySelector('.creationDate').innerText = bookmark.CreationDate;
            tbody.appendChild(temp);


        }

    }
    if (document.getElementById("searchShortReview").innerText.length === 0) {
        searchShortReview();
    }
    let select = document.getElementById("filterBookmarks");
    if (select.selectedIndex !== 0) {
        filterBookmarks();
    }

}

function filterBookmarks() {
    let select = document.getElementById("filterBookmarks");

    let index = select.selectedIndex;
    let options = select.options;
    let filter = options[index].innerText.toUpperCase();
    let tbody = document.getElementById('tbody');
    let tr = tbody.getElementsByTagName('tr');
    //if index===0 no filter is selected
    if (index !== 0) {
        for (let i in tr) {
            try {
                let td = tr[i].getElementsByTagName('td')[5];
                if (td) {
                    let txtValue = td.textContent || td.innerText;
                    if (txtValue.toUpperCase().indexOf(filter) > -1) {
                        tr[i].style.display = "";
                    } else {
                        tr[i].style.display = "none";
                    }
                }
            } catch (e) {

            }

        }
    } else {
        for (let i in tr) {
            try {
                let td = tr[i].getElementsByTagName('td')[5];
                if (td) {
                    tr[i].style.display = "";

                }
            } catch (e) {

            }

        }
    }


}


function searchShortReview() {
    let tbody = document.getElementById('tbody');
    let tr = tbody.getElementsByTagName('tr');
    let input = document.getElementById('searchShortReview');
    let filter = input.value.toUpperCase();


    for (let i in tr) {
        try {
            let td = tr[i].getElementsByTagName('td')[1];
            if (td) {
                let txtValue = td.textContent || td.innerText;

                if (txtValue.toUpperCase().indexOf(filter) > -1) {
                    tr[i].style.display = "";
                } else {
                    tr[i].style.display = "none";
                }
            }
        } catch (e) {

        }

    }

}


