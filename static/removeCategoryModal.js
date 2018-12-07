

function openRemoveCategoryModal(e){
    //die url von dem die kategorie entfernt werden soll
    let catName=e.target;
    let id=catName.parentElement.previousSibling.previousSibling.id;
    let div=document.getElementById('container');
    let p=document.createElement('p');
    let button=document.createElement("button");
    button.innerText="bestätigen";
    button.addEventListener('click',()=>{
        let xhr=new XMLHttpRequest();
        let url='removeCategory';
        let formData=new FormData();
        formData.append('url',id);
        formData.append('category',catName.innerText);
        xhr.open('POST',url);
        xhr.send(formData);


    });
    p.innerText="Wollen Sie "+catName.innerText+" wirklich entfernen?";
    div.appendChild(p);
    div.appendChild(button);
    removeCategoryModal.style.display = "block";


}
// Get the modal
let removeCategoryModal = document.getElementById('removeCategoryModal');

// Get the button that opens the modal
newCategoryButton = document.getElementById("newCategory");

// Get the <span> element that closes the modal

// When the user clicks anywhere outside of the modal, close it
window.addEventListener('click',function(event) {
    console.log(event.target==removeCategoryModal);
    if (event.target == removeCategoryModal) {
        removeCategoryModal.style.display = "none";
    }
},false);
