
function openNewCategoryModal(){
    console.log('openNewCategoryModal');
    newCategoryModal.style.display = "block";
}
let sendButton=document.getElementById("send");
sendButton.addEventListener("click",(e)=>{
    e.preventDefault();
    let formData=new FormData();
    let catName=document.getElementById("catName").value;
    formData.append("catName",catName);
    console.log(formData)
    var xhr=new XMLHttpRequest();
    xhr.addEventListener("load",()=>{
        console.log(xhr.responseText)
    });
    xhr.open('POST',"newCategory");
    xhr.send(formData);
    newCategoryModal.style.display = "none";
});
// Get the modal
let newCategoryModal = document.getElementById('newCategoryModal');



// When the user clicks anywhere outside of the modal, close it

window.addEventListener('click',function(event) {
    console.log(event.target==newCategoryModal);
    if (event.target == newCategoryModal) {
        newCategoryModal.style.display = "none";
    }
},false);