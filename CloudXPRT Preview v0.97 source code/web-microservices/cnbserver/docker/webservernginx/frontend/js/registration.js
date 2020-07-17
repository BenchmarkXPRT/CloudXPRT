async function register(form) {
	var username		= document.getElementById("username").value;
	var password		= document.getElementById("password").value;
	var verifyPassword 	= document.getElementById("verifyPassword").value;
	var email 			= document.getElementById("email").value;
	var verifyEmail 	= document.getElementById("verifyEmail").value;

	if (username == "" ||
		password == "" ||
		verifyPassword == "" ||
		email == "" ||
		verifyEmail == "") {

		alert("Please fill in all fields");
		return;
	}

	if (!validPassword(password)) {
		alert("Password does not meet minimum requirements:\n7-20 characters consisting of atleast:\none uppercase letter,\none lowercase letter,\none number,\nand one of the following special characters (!,#,$,%)");
		return;
	}

	if (!validEmail(email)) {
		alert("Enter a valid email address");
		return;	
	}

	if (password == username) {
		alert("Password cannot be the same as your user ID");
		return;
	}

	if (password != verifyPassword) {
		alert("Your passwords entries do not match");
		return;
	}

	if (email != verifyEmail) {
		alert("Your email entries do not match");
		return;
	}

	// Call REST API for registration
	const data = { 
		id: username, 
		password: password,
		email: email
	}

	url = 'http://' + location.host + '/users'
	const response = await fetch(url, {
		method: 'POST',
		body: JSON.stringify(data),
		headers: {
			'Content-Type': 'application/json'
		}
	});

	alert("Thank you for registering!")
	clearForm(form);
	window.location = "index.html"
}

function validPassword(password) {
	var re = /^(?=.*?[A-Z])(?=.*?[a-z])(?=.*?[0-9])(?=.*[!#$%]).{7,20}/;
	return re.test(String(password));
}

function validEmail(email) {
    var re = /^(([^<>()\[\]\\.,;:\s@"]+(\.[^<>()\[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/;
    return re.test(String(email).toLowerCase());
}

function clearForm(form) {
	document.getElementById(form).reset();
	window.location = "index.html"
}

document.onkeyup = function (event) {
	if (event.keyCode === 13) {
		event.preventDefault();
		document.getElementById("register-button").click();
	}
}