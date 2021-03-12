var passwordElement	= document.getElementById("password");

passwordElement.addEventListener('keyup', function() {
    var password = passwordElement.value;

    if (password.length == 0) {
        document.getElementById("progress").value = "0";
        return;
    }

    //regular expressions for password strength criteria
    var prog = [/[!#$%]/, /[A-Z]/, /[0-9]/, /[a-z]/].reduce((memo, test) => memo + test.test(password), 0);

    if (password.length >= 7 && password.length <= 20) {
        prog++;
    }

    var progress = "";
    switch (prog) {
        case 0:
        case 1:
        case 2:
            progress = "25";
            break;
        case 3:
            progress = "50";
            break;
        case 4:
            progress = "75";
            break;
        case 5:
            progress = "100";
            break;
    }

    document.getElementById("progress").value = progress;
});