  /**
   * Scroll spy for existing table of contents that highlights items based on visible sections
   */

  // Configuration options
  const config = {
    // Selector for the headings to track (e.g., h2, h3, etc.)
    headingSelector: "h1, h2, h3, h4, h5, h6",

    // Selector for the table of contents links
    tocLinkSelector: "#table-of-contents .tocNode",

    // Class to add to TOC items when their section is active/visible
    activeClass: "active",

    // How much of the section needs to be visible to be considered active (0-1)
    intersectionThreshold: 0.4,

    // Offset from the top of the viewport (in pixels)
    offsetTop: 100,
  };

  // Initialize the scroll spy
  function initScrollSpy() {
    // Get all the headings we want to track
    const headings = document.querySelectorAll(config.headingSelector);

    // Get the TOC links
    const tocLinks = document.querySelectorAll(config.tocLinkSelector);

    // Whose scroll do we observe?
    const root = document.querySelector("#contentPanel");

    // Exit if either headings or TOC links don't exist
    if (!headings.length || !tocLinks.length) {
      console.error("Could not find headings or table of contents links");
      return;
    }

    // Create a mapping between heading IDs and TOC links
    const idToLinkMap = Array.from(tocLinks).reduce((map, link) => {
      // Get the target ID from the href attribute (remove the leading #)
      const targetId = link.getAttribute("href").replace(/^#/, "");
      map[targetId] = link;
      return map;
    }, {});

    // Set up the Intersection Observer to detect when sections are visible
    setupIntersectionObserver2(root, headings, idToLinkMap);
  }

  // Set up the intersection observer to watch headings
  function setupIntersectionObserver1(root, headings, idToLinkMap) {
    const tocLinks = document.querySelectorAll(config.tocLinkSelector);

    // Options for the Intersection Observer
    const options = {
      root: root,
      rootMargin: `-${config.offsetTop}px 0px 0px 0px`,
      threshold: config.intersectionThreshold,
    };

    // Create the observer
    const observer = new IntersectionObserver((entries) => {
      // Store visible sections
      const visibleSections = [];

      entries.forEach((entry) => {
        // Get the ID of the section
        const id = entry.target.id;

        // If the section is visible, add it to our array
        if (entry.isIntersecting) {
          visibleSections.push({
            id: id,
            // We'll use this to determine which section is "most visible"
            visibleRatio: entry.intersectionRatio,
          });
        }
      });

      // Clear active class from all TOC links
      tocLinks.forEach((link) => {
        link.classList.remove(config.activeClass);

        // If the link is in a list item, also remove the class from the parent
        if (link.parentElement.tagName === "LI") {
          link.parentElement.classList.remove(config.activeClass);
        }
      });

      // If we have visible sections, highlight the most visible one
      if (visibleSections.length > 0) {
        // Sort by visibility ratio (most visible first)
        visibleSections.sort((a, b) => b.visibleRatio - a.visibleRatio);

        // Get the TOC link for the most visible section
        const mostVisibleId = visibleSections[0].id;
        const activeLink = idToLinkMap[mostVisibleId];

        // Highlight the active TOC link
        if (activeLink) {
          activeLink.classList.add(config.activeClass);

          // If the link is in a list item, also add the class to the parent
          if (activeLink.parentElement.tagName === "LI") {
            activeLink.parentElement.classList.add(config.activeClass);
          }
        }
      }
    }, options);

    // Start observing each heading
    headings.forEach((heading) => {
      // Make sure the heading has an ID
      if (!heading.id) {
        console.warn(
          `Heading without ID found: ${heading.textContent}. Skipping.`,
        );
        return;
      }

      observer.observe(heading);
    });

    // Return the observer in case we need to disconnect it later
    return observer;
  }

  // Set up the intersection observer to watch headings
  function setupIntersectionObserver2(root, headings, idToLinkMap) {
    const tocLinks = document.querySelectorAll(config.tocLinkSelector);
    
    // Options for the Intersection Observer
    const options = {
      root: root,
      rootMargin: `-${config.offsetTop}px 0px 0px 0px`,
      threshold: config.intersectionThreshold
    };
    
    // Create the observer
    const observer = new IntersectionObserver((entries) => {
      // Process each entry
      entries.forEach(entry => {
        // Get the ID of the section
        const id = entry.target.id;
        // Get the TOC link for this section
        const link = idToLinkMap[id];
        
        if (!link) return; // Skip if no matching link found
        
        // When section enters viewport
        if (entry.isIntersecting) {
          // Add active class to the TOC link
          link.classList.add(config.activeClass);
          
          // If the link is in a list item, also add the class to the parent
          if (link.parentElement.tagName === 'LI') {
            link.parentElement.classList.add(config.activeClass);
          }
        } 
        // When section leaves viewport
        else {
          // Remove active class from the TOC link
          link.classList.remove(config.activeClass);
          
          // If the link is in a list item, also remove the class from the parent
          if (link.parentElement.tagName === 'LI') {
            link.parentElement.classList.remove(config.activeClass);
          }
        }
      });
      
      // Optional: if you want to ensure at least one item is always highlighted
      // even if no sections are properly visible, uncomment this block:
      /*
      const anyActive = Array.from(tocLinks).some(link => 
        link.classList.contains(config.activeClass)
      );
      
      if (!anyActive && tocLinks.length > 0) {
        // Get the current scroll position
        const scrollPos = window.scrollY;
        
        // Find the nearest heading to the current scroll position
        let closestHeading = null;
        let closestDistance = Infinity;
        
        headings.forEach(heading => {
          const distance = Math.abs(heading.getBoundingClientRect().top + scrollPos - scrollPos);
          if (distance < closestDistance) {
            closestDistance = distance;
            closestHeading = heading;
          }
        });
        
        // Highlight the TOC link for the closest heading
        if (closestHeading && idToLinkMap[closestHeading.id]) {
          const link = idToLinkMap[closestHeading.id];
          link.classList.add(config.activeClass);
          
          if (link.parentElement.tagName === 'LI') {
            link.parentElement.classList.add(config.activeClass);
          }
        }
      }
      */
    }, options);
    
    // Start observing each heading
    headings.forEach(heading => {
      // Make sure the heading has an ID
      if (!heading.id) {
        console.warn(`Heading without ID found: ${heading.textContent}. Skipping.`);
        return;
      }
      
      observer.observe(heading);
    });
    
    // Return the observer in case we need to disconnect it later
    return observer;
  }

  // Re-initialize when the window is resized (debounced)
  let resizeTimer;
  window.addEventListener("resize", () => {
    clearTimeout(resizeTimer);
    resizeTimer = setTimeout(initScrollSpy, 250);
  });
