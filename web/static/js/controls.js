document.getElementById('file').onchange = () => {
    if (document.getElementById('file').value) {
        document.getElementById('submit').disabled = false;
    }
};
