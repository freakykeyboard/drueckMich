'use strict';
//ToDo hasOwnProperty
window.addEventListener("load", () => {

    setInterval(updateBookmarks, 1000 * 5);

    function updateBookmarks() {
        let xhr = new XMLHttpRequest();
        const url = "update";

        xhr.addEventListener("load", () => {
            updateContent(JSON.parse(xhr.responseText));
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

    });
    xhr.open("POST", url);

    formData.append('orderBy', "0");
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
    if (data.available_categories){
        let select=document.getElementById('filterBookmarks');

        while (select.firstChild.nextSibling) {
            select.removeChild(select.firstChild);
        }
        select.removeChild(select.firstChild);
        let temp=document.getElementById('selectCategory').content.cloneNode(true);
        temp.querySelector('.category').innerText="nach Kategorien filtern";

        select.appendChild(temp);
        for (let i in data.available_categories){
            let category=data.available_categories[i];
            let temp=document.getElementById('selectCategory').content.cloneNode(true);
            temp.querySelector('.category').innerText=category;

            select.appendChild(temp);
        }
    }
    if (data.bookmarks) {
        for (let i = 0; i < data.bookmarks.length; i++) {
            let bookmark=data.bookmarks[i];

            let temp=document.getElementById('bookmarkRow').content.cloneNode(true);
            temp.querySelector('.url').innerText=bookmark.url;
            temp.querySelector('.shortReview').innerText=bookmark.shortReview;
            temp.querySelector('.title').innerText=bookmark.title;
            for (let j =0;j<bookmark.images.length;j++){
                let li=document.createElement('li');
                li.innerText=bookmark.images[j];
                temp.querySelector('.images').appendChild(li);
            }
            for (let j in bookmark.wvr_categories){
                let li=document.createElement('li');
                li.innerText=bookmark.wvr_categories[j];
                temp.querySelector('.wvrCategories').appendChild(li);
            }
            for (let j in bookmark.custom_categories){
                let li=document.createElement('li');
                li.innerText=bookmark.custom_categories[j];
                temp.querySelector('.customCategories').appendChild(li);
            }
            temp.querySelector('.imgIcon').setAttribute('src',"/gridGetIcon?fileName="+bookmark.icon);

            temp.querySelector('.lat').innerText+=bookmark.lat;
            temp.querySelector('.lon').innerText+=bookmark.long;
            //ToDo saving id instead of url?
             temp.querySelector('.addCategory').setAttribute('id',bookmark.url);
             temp.querySelector('.customCategories').setAttribute("id",bookmark.url);
            tbody.appendChild(temp);


        }

    }
}
function filterBookmarks(e){
    console.log('filterBookmarks');
    let index=e.target.selectedIndex;
    let options=e.target.options;
    let filter=options[index].innerText.toUpperCase();
   let tbody=document.getElementById('tbody');
   let tr=tbody.getElementsByTagName('tr');

   for (let i in tr){
       console.log(tr);
       let td=tr[i].getElementsByTagName('td')[6];
       console.log(td);
       if (td){
           let txtValue=td.textContent||td.innerText;
           if (txtValue.toUpperCase().indexOf(filter)>-1){
               tr[i].style.display="";
           } else {
               tr[i].style.display="none";
           }
       }
   }


}

