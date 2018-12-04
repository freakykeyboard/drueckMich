

let sendButton=document.getElementById("send");
sendButton.addEventListener("click",(e)=>{
    e.preventDefault()
    let formData=new FormData();
    let catName=document.getElementById("catName").value;
    formData.append("catName",catName);
    console.log(formData)
    var xhr=new XMLHttpRequest();
    xhr.addEventListener("load",()=>{
        console.log(xhr.responseText)
    });
    xhr.open('POST',"newCategory");
    xhr.send(formData)
    modal.style.display = "none";
});
// Get the modal
var modal = document.getElementById('myModal');

// Get the button that opens the modal
var btn = document.getElementById("newCategory");

// Get the <span> element that closes the modal


// When the user clicks the button, open the modal
btn.onclick = function() {
    modal.style.display = "block";
}



// When the user clicks anywhere outside of the modal, close it
window.onclick = function(event) {
    if (event.target == modal) {
        modal.style.display = "none";
    }
}