// parceimg add to all img - event - on mouse enter 
// on mouse enter - will replace src on original image src taken from parent href
// pattern: <a href="14_.png"><img src="14.png"></a>
function parceimg() {
    var images = document.querySelectorAll("img");
    for (var i = 0; i < images.length; i++) {
        if (images[i].parentElement.tagName.toLowerCase() === "a") {
            images[i].onmouseenter = function(el) {
                var href = this.parentElement.getAttribute("href");
                var src = this.getAttribute("src");
                //already replaced?
                if (href!=src) {
                    // check href/src
                    if (href!=null && src!=null && href.length>5 && src.length>5 && href.length - src.length ==1) {
                        var hrefMust = src.substr(0,src.length-4) + "_" + src.substr(src.length-4)
                        if (hrefMust == href) {
                            // replace
                            this.src = href;
                        }
                    }
                }
            }
            images[i].onmouseleave = function(el) {
                var href = this.parentElement.getAttribute("href");
                var src = this.getAttribute("src");
                if (href==src) {
                    var hrefMust = src.substr(0,src.length-5) + src.substr(src.length-4)
                    //console.log(hrefMust)
                    this.src = hrefMust;
                }
            }
        }
    }  
}
window.onload = function() {
    parceimg();
}
