

function openRemoveCategoryModal(e){
    //die url von dem die kategorie entfernt werden soll
    id=e.target.getAttribute("id");
let div=document.getElementById("container");
let xhr=new XMLHttpRequest();
let url="update";
    xhr.addEventListener('load', () => {
        let response = JSON.parse(xhr.responseText);
        console.log(response);
        let temp,item,a,p,i,item2;
        //get the tenmplate element
        temp = document.getElementsByTagName("template")[2];
        item = temp.content.querySelector("div");
        item2=item.querySelector("p");
        console.log(item);
        for (i in response.bookmarks) {
            if (response.bookmarks[i].url===id){
                for (let j in response.bookmarks[i].custom_categories) {
                    a = document.importNode(item, true);
                    p=document.importNode(item2,true);
                    a.textContent +=response.bookmarks[i].custom_categories[j];

                    p.innerHTML="&#9746";
                    p.setAttribute("id",id);
                    div.appendChild(a);

                    a.appendChild(p);

                                  }


            }

        }
    });
    // When the user clicks the button, open the modal
    xhr.open("GET",url);
    xhr.send();
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
function removeCategory(e){
    console.log('removeCategoryHandler');
    console.log(e.target)
}