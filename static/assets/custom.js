function Testing(message) {
    const html = `<h1>${JSON.stringify(message)}</h1>`
    let elem = document.getElementById("default-sidebar")
    elem.innerHTML = html + elem.innerHTML
}