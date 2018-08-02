;(function(){
    'use strict';

    var app = {};

    app.images = {
        handlers: function(node, url) {
            node.addEventListener('mouseenter', function() {
                this.setAttribute('data-compressed', this.src);
                this.src = url;
            })

            node.addEventListener('mouseleave', function() {
                this.setAttribute('data-original', this.src);
                this.src = this.getAttribute('data-compressed');
            })
        },

        init: function() {
            var images = document.querySelectorAll('article section img');

            for (var i = 0; images.length > i; i++) {
                var current = images[i];
                var parent = current.parentElement;
                var parentTagname = parent.tagName.toLowerCase();

                if (parentTagname === 'a') {
                    var normalImage = parent.getAttribute('href');
                    var img = new Image();

                    img.onload = function() { app.images.handlers(current, normalImage) }
                    img.src = normalImage;
                }
            }
        }
    }

    window.onload = function() {
        for(var o in app) { app[o].init(); }
    }
}());