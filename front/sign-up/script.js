document.addEventListener('DOMContentLoaded', function() {
    document.querySelector('form').addEventListener('submit', function(e) {
        e.preventDefault();
    });
    document.getElementById('submitBtn').addEventListener('click', buttonHandler);

});

async function buttonHandler() {
    const success=document.getElementById('successMessage');
    let fullNameInput = document.getElementById('fullName');
    let emailInput = document.getElementById('email');
    let passwordInput = document.getElementById('password');
    // let passwordAgainInput=document.getElementById('confirmPassword');
    let countdown=document.getElementById('countdown');
    console.log(fullNameInput.value,emailInput.value,passwordInput.value,passwordInput.value);
    if (fullNameInput.value===""||emailInput.value===""|| passwordInput.value==="") {
        console.log("null fields")
        throw new Error("null fields");

    }
    const data={
        full_name:fullNameInput.value,
        username:emailInput.value,
        password:passwordInput.value,
    }
    console.log(data.full_name)
     try{
         const response= await fetch("http://localhost:8284/auth/sign-up",{
             method:"POST",
             headers:{
                 'Content-Type': 'application/json',
             },
             body:JSON.stringify(data)
         })
         if (!response.ok){
             if (response.status === 422){
                 success.classList.remove('hidden');
                 success.children[0].children[0].removeAttribute('class')
                 success.children[0].children[0].setAttribute('class','fa-solid fa-question text-sky-600 text-2xl')
                 success.children[0].children[1].textContent="User with the same email already exists";
                 success.children[1].textContent="Seems like you are already registered"
                 success.children[2].textContent="Redirect To Login in 3 seconds"

                 setTimeout(() => {
                     window.location.href="../sign-in/sign-in.html";
                 }, 3000); // 3000ms = 3 seconds
             }else{
                 console.log("error")
                 throw new Error("http error");
             }

         }
         const result=await response.json();
         console.log("success",result);

         success.classList.remove('hidden');




     }catch(err){
         console.log(err);
     }
}