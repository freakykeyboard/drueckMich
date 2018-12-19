

// Get the modal
let registrationModal = document.getElementById('registrationModal');

// Get the button that opens the modal
let btn = document.getElementById("registrationButton");

document.getElementById('registrate').addEventListener('click',()=>{
    registrationModal.style.display = "none";

});


// When the user clicks the button, open the modal
btn.onclick = function() {
    registrationModal.style.display = "block";
};



// When the user clicks anywhere outside of the modal, close it
window.addEventListener('click',function(event) {
    console.log(event.target==registrationModal);
    if (event.target == registrationModal) {

        registrationModal.style.display = "none";

    }
},false);