let id;
function addCategory(e) {
    let category=e.target.value;
    let xhr=new XMLHttpRequest();
    let url='addCategoryToBookmark';
    xhr.addEventListener('load',()=>{
    updateTable(JSON.parse(xhr.responseText));
    });
    let formData=new FormData();
    formData.append('url',id);
    formData.append('category',category);
    xhr.open('POST',url);
    xhr.send(formData);
}
function addCategoryToBookmark(e) {
    console.log('addCategoryModal');
    id=e.target.getAttribute("id");
    console.log(id);
let select=document.getElementById("addSelect");
    let xhr = new XMLHttpRequest();
    let url = "update";
    xhr.addEventListener('load', () => {
        let response = JSON.parse(xhr.responseText);

        let temp,item,a,i;
        //get the tenmplate element
        temp = document.getElementsByTagName("template")[1];
        item = temp.content.querySelector("option");
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