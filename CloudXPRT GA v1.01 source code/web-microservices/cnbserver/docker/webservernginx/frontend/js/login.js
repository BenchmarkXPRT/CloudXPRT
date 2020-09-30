async function login() {
	var username = document.getElementById("username").value;
	var password = document.getElementById("password").value;

	if (username == "" ||
		password == "") {

		alert("Please enter username and password to login");
		return;
	}

	// Call REST API to verify log in credentials
	url = 'http://' + location.host + '/users/' + username

	const response = await fetch(url, {
		method: 'GET',
		headers: {
			'Content-Type': 'application/json'
		}
	})
	.then(response => response.json())
	.then(data => {
		if (data.id == username && data.password == password) {
			window.localStorage.setItem("username", username)
			window.location = 'options.html'
		} else {
			alert("Incorrect username/password")
		}
	})
	.catch(err => {
		alert("User does not exist");
	});
}

document.onkeyup = function (event) {
	if (event.keyCode === 13) {
		event.preventDefault();
		document.getElementById("signin-button").click();
	}
}