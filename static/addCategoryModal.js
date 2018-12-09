let id;
function addCategory(e) {
    let category=e.target.value;
    let xhr=new XMLHttpRequest();
    let url='addCategoryToBookmark';
    xhr.addEventListener('load',()=>{
    updateContent(JSON.parse(xhr.responseText));
    });
    let formData=new FormData();
    formData.append('url',id);
    formData.append('category',category);
    xhr.open('POST',url);
    xhr.send(formData);
    addCategoryModal.style.display = "none";
}
function addCategoryToBookmark(e) {
    console.log('addCategoryModal');
    console.log(e.target.previousSibling);
    id=e.target.getAttribute("id");
    console.log(id);
let select=document.getElementById("addSelect");
    let xhr = new XMLHttpRequest();
    let url = "update";
    xhr.addEventListener('load', () => {
        while (select.firstChild.nextSibling) {

            select.removeChild(select.firstChild);
        }
        select.removeChild(select.firstChild);
        let response = JSON.parse(xhr.responseText);

        let temp,item,a,i;
        //get the tenmplate element
        temp = document.getElementById('selectMenu');
        item = temp.content.querySelector("option");
        a = document.importNode(item, true);
        a.innerText="KAategorie auswählen";
        select.appendChild(a);
        for (i in response.available_categories) {
            let category = response.available_categories[i];
            a = document.importNode(item, true);
            a.textContent = category;

            select.appendChild(a);
        }

    });
    // When the user clicks the button, open the modal
    xhr.open("GET",url);
    xhr.send();
    addCategoryModal.style.display = "block";
}
// Get the modal
let addCategoryModal = document.getElementById('addCategoryModal');


window.addEventListener('click',function(event) {
    console.log(event.target==addCategoryModal);
    if (event.target == addCategoryModal) {
        addCategoryModal.style.display = "none";
    }
},false);