

let sendButton=document.getElementById("registrate");
sendButton.addEventListener("click",(e)=>{
    e.preventDefault()
    let formData=new FormData();
    let userName=document.getElementById("userName").value;
    let password=document.getElementById("password").value;
    formData.append("username",userName);
    formData.append("password",password);
    let xhr=new XMLHttpRequest();
    xhr.addEventListener("load",()=>{
        console.log(xhr.responseText)
    });
    xhr.open('POST',"registrate");
    xhr.send(formData);
    registrationModal.style.display = "none";
});
// Get the modal
let registrationModal = document.getElementById('registrationModal');

// Get the button that opens the modal
let btn = document.getElementById("registrationButton");

// Get the <span> element that closes the modal


// When the user clicks the button, open the modal
btn.onclick = function() {
    registrationModal.style.display = "block";
}



// When the user clicks anywhere outside of the modal, close it
window.addEventListener('click',function(event) {
    console.log(event.target==registrationModal);
    if (event.target == registrationModal) {

        registrationModal.style.display = "none";

    }
},false);