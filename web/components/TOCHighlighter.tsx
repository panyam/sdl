
export class Config {
  // Selector for the headings to track (e.g., h2, h3, etc.)
  headingSelector = "h1, h2, h3, h4, h5, h6";

  // Selector for the table of contents links
  tocLinkSelector = ".tocNode";

  // Class to add to TOC items when their section is active/visible
  activeClass = "active"

  // How much of the section needs to be visible to be considered active (0-1)
  intersectionThreshold = 0.4;

  // Offset from the top of the viewport (in pixels)
  offsetTop = 100;

  // Auto-scroll the TOC to keep active items visible
  autoScrollTOC = true;

  // Padding (in pixels) to add at the top and bottom when scrolling the TOC
  tocScrollPadding = 20;
}

export default class TOCHighlighter {
  config: Config
  resizeTimer: any;
  constructor(public readonly contentRoot: HTMLDivElement, public readonly tocRoot: HTMLDivElement, config?: Config) {
    config = config || new Config()
    this.config = config
    const headings = contentRoot.querySelectorAll(config.headingSelector)
    const tocLinks = tocRoot.querySelectorAll(config.tocLinkSelector)

    // Exit if either headings or TOC links don't exist
    if (!headings.length || !tocLinks.length) {
      console.error("Could not find headings or table of contents links");
      return;
    }

    // Create a mapping between heading IDs and TOC links
    const idToLinkMap = Array.from(tocLinks).reduce((map: any, link) => {
      // Get the target ID from the href attribute (remove the leading #)
      const targetId = ((link.getAttribute("href") as string) || "").replace(/^#/, "");
      map[targetId] = link;
      return map;
    }, {});
    this.setupIntersectionObserver2(headings, idToLinkMap)

    window.addEventListener("resize", () => {
      clearTimeout(this.resizeTimer);
      this.resizeTimer = setTimeout(() => {
        this.setupIntersectionObserver2(headings, idToLinkMap)
      }, 250);
    });
  }

  // Set up the intersection observer to watch headings
  setupIntersectionObserver1(headings: NodeListOf<Element>, idToLinkMap: any) {
    const config = this.config;
    const tocLinks = document.querySelectorAll(config.tocLinkSelector);

    // Options for the Intersection Observer
    const options = {
      root: this.contentRoot,
      rootMargin: `-${config.offsetTop}px 0px 0px 0px`,
      threshold: config.intersectionThreshold,
    };

    // Create the observer
    const observer = new IntersectionObserver((entries) => {
      // Store visible sections
      const visibleSections = [] as any[];

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
      tocLinks.forEach((link: Element) => {
        link.classList.remove(config.activeClass);

        // If the link is in a list item, also remove the class from the parent
        if (link.parentElement && link.parentElement.tagName === "LI") {
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
  setupIntersectionObserver2(headings: NodeListOf<Element>, idToLinkMap: any) {
    const config = this.config;
    const tocLinks = document.querySelectorAll(config.tocLinkSelector);
    
    // Options for the Intersection Observer
    const options = {
      root: this.contentRoot,
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
          
          // Auto-scroll the TOC to ensure the active item is visible
          if (config.autoScrollTOC) {
            this.scrollTOCToActiveItem(link);
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
      const anyActive = Array.from(tocLinks).some(link => 
        link.classList.contains(config.activeClass)
      );
      
      if (!anyActive && tocLinks.length > 0) {
        // Get the current scroll position
        const scrollPos = window.scrollY;
        
        // Find the nearest heading to the current scroll position
        let closestHeading = null as any;
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
    }, options);
    
    // Start observing each heading
    headings.forEach((heading: any) => {
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

  // Function to scroll the TOC container to make an active item visible
  scrollTOCToActiveItem(activeLink: any) {
    const config = this.config;
    const tocContainer = this.tocRoot; // document.querySelector(config.tocContainerSelector);

    // Only proceed if we have a scrollable container
    if (!tocContainer || !activeLink) return;

    // Check if the container is scrollable (has overflow)
    const containerStyle = window.getComputedStyle(tocContainer);
    const isScrollable = ['auto', 'scroll'].includes(containerStyle.overflowY) ||
                         tocContainer.scrollHeight > tocContainer.clientHeight;

    if (!isScrollable) return;

    // Get positions
    const containerRect = tocContainer.getBoundingClientRect();
    const linkRect = activeLink.getBoundingClientRect();

    // Calculate relative positions
    const linkTop = linkRect.top - containerRect.top;
    const linkBottom = linkRect.bottom - containerRect.top;

    // Check if the link is outside the visible area of the container
    if (linkTop < config.tocScrollPadding) {
      // Link is above the visible area, scroll up to show it
      tocContainer.scrollTop += (linkTop - config.tocScrollPadding);
    } else if (linkBottom > containerRect.height - config.tocScrollPadding) {
      // Link is below the visible area, scroll down to show it
      tocContainer.scrollTop += (linkBottom - containerRect.height + config.tocScrollPadding);
    }

    // Alternate smoother approach using scrollIntoView (commented out by default)
    /*
    activeLink.scrollIntoView({
      behavior: 'smooth',
      block: 'nearest',
      inline: 'nearest'
    });
    */
  }
}
