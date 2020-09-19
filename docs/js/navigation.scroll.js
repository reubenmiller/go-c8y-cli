---
layout: null
---
(function () {

    function getNavigationElement() {
        var el = document.getElementById('navbar');
        return el;
    }

    function isOnScreen(element) {
        var position = element.getBoundingClientRect();

        // checking whether fully visible
        if(position.top >= 0 && position.bottom <= window.innerHeight) {
            return true;
        }

        // checking for partial visibility
        if(position.top < window.innerHeight && position.bottom >= 0) {
            return true;
        }
        return false;
    }

    function scrollToSelectedItemInNavBar() {
        var el = document.querySelector('.current');

        if (el) {
            if (!isOnScreen(el)) {
                el.scrollIntoView(true);
            }
        }
    }

    document.addEventListener("DOMContentLoaded", function (event) {
        var scrollpos = sessionStorage.getItem('scrollpos');
        if (scrollpos) {
            value = parseInt(scrollpos, 10);

            var el = getNavigationElement()
            if (el) {
                el.scrollTop = value;
            }

            if (value === 0) {
                scrollToSelectedItemInNavBar();
            }

            sessionStorage.removeItem('scrollpos');
        }
    });

    window.addEventListener("beforeunload", function (e) {
        var el = getNavigationElement();

        if (el) {
            sessionStorage.setItem('scrollpos', el.scrollTop);
        }
    });
})();
