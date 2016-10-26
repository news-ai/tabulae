var codeParameter = $.urlParam('code');
if (codeParameter) {
    document.getElementById("invitation_code").value = codeParameter;
}

var emailParameter = $.urlParam('email');
if (emailParameter) {
    document.getElementById("email").value = emailParameter;
}