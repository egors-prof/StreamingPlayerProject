document.addEventListener('DOMContentLoaded', function() {
    document.querySelector('form').addEventListener('submit', function(e) {
        e.preventDefault();
    });
    document.getElementById('submitBtn').addEventListener('click', buttonHandler);

});


async function buttonHandler() {
    const success=document.getElementById('successMessage');
    let emailInput = document.getElementById('email');
    let passwordInput = document.getElementById('password');
    // let passwordAgainInput=document.getElementById('confirmPassword');
    let countdown=document.getElementById('countdown');
    console.log(emailInput.value, passwordInput.value);
    if (emailInput.value===""|| passwordInput.value==="") {
        console.log("null fields")
        throw new Error("null fields");

    }
    const data={
        username:emailInput.value,
        password:passwordInput.value,
    }

    try{
       const response = await fetch('http://localhost:8284/auth/sign-in',{
           method:"POST",
           headers:{
               'Content-Type': 'application/json',
           },
           body:JSON.stringify(data)
       })
        const result= await response.json()
        console.log(result.access_token,result.refresh_token)
        localStorage.setItem('access_token',result.access_token)
        localStorage.setItem('refresh_token',result.refresh_token)



    }catch(err){
        console.log(err);
    }
}