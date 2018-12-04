let sendButton1 = document.getElementById("send");
sendButton1.addEventListener("click", (e) => {
    e.preventDefault()
    let formData = new FormData();
    let catName = document.getElementById("catName").value;
    formData.append("catName", catName);
    console.log(formData)
    var xhr = new XMLHttpRequest();
    xhr.addEventListener("load", () => {
        console.log(xhr.responseText)
    });
    xhr.open('POST', "newCategory");
    xhr.send(formData);
    addCategoryModal.style.display = "none";
});
// Get the modal
let addCategoryModal = document.getElementById('addCategoryModal');


function addCategoryToBookmark() {
    let select = document.getElementById("addSelect");
    let xhr = new XMLHttpRequest();
    let url = "update";
    xhr.addEventListener('load', () => {
        let response = JSON.parse(xhr.responseText);
        console.log(response)
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


// When the user clicks anywhere outside of the modal, close it
window.onclick = function (event) {
    if (event.target == addCategoryModal) {
        addCategoryModal.style.display = "none";
    }
}