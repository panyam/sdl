
class TOCHierarchy {
  // Map of child -> parent links
  parentMap = new Map();
  // Map of link -> all ancestor links
  ancestorMap = new Map();
  // Map of link -> all descendant links
  descendantMap = new Map();
  // Heading level map (h2, h3, h4, etc.)
  levelMap = new Map();
}

export class Config {
  // Selector for the headings to track (e.g., h2, h3, etc.)
  headingSelector = "h1, h2, h3, h4, h5, h6";

  // Selector for the table of contents links
  tocLinkSelector = ".tocNode";

  // Class to add to TOC items when their section is active/visible
  activeClass = "active"

  // Class to add to parent/ancestor items of active items
  parentActiveClass = 'parent-active';

  // How much of the section needs to be visible to be considered active (0-1)
  intersectionThreshold = 0.5;

  // Offset from the top of the viewport (in pixels)
  offsetTop = 100;

  // Auto-scroll the TOC to keep active items visible
  autoScrollTOC = true;

  // Padding (in pixels) to add at the top and bottom when scrolling the TOC
  tocScrollPadding = 20;

  // Respect hierarchy when highlighting items
  respectHierarchy = true;
}

export default class TOCHighlighter {
  config: Config
  resizeTimer: any;
  constructor(public readonly contentRoot: HTMLDivElement, public readonly tocRoot: HTMLDivElement, config?: Config) {
    config = config || new Config()
    this.config = config

    this.initialize()

    window.addEventListener("resize", () => {
      clearTimeout(this.resizeTimer);
      this.resizeTimer = setTimeout(() => {
        this.initialize();
      }, 250);
    });
  }

  initialize() {
    const config = this.config;
    const headings = this.contentRoot.querySelectorAll(config.headingSelector)
    const tocLinks = this.tocRoot.querySelectorAll(config.tocLinkSelector)

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

    // If we respect hierarchy, build a map of TOC parent-child relationships
    let tocHierarchy = null;
    if (config.respectHierarchy) {
      tocHierarchy = this.buildTOCHierarchy(tocLinks);
    }

    // Set up the Intersection Observer to detect when sections are visible
    this.setupIntersectionObserver(headings, idToLinkMap, tocHierarchy)
  }

  // Set up the intersection observer to watch headings
  setupIntersectionObserver(headings: NodeListOf<Element>, idToLinkMap: any, tocHierarchy: TOCHierarchy | null) {
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

          // Handle nested TOC structure if hierarchy support is enabled
          if (config.respectHierarchy && tocHierarchy) {
            // Add parent-active class to all ancestors
            const ancestors = tocHierarchy.ancestorMap.get(link) || [];
            ancestors.forEach((ancestor: any) => {
              ancestor.classList.add(config.parentActiveClass);

              if (ancestor.parentElement.tagName === 'LI') {
                ancestor.parentElement.classList.add(config.parentActiveClass);
              }
            });
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

  // Build a hierarchical map of the TOC structure
  buildTOCHierarchy(tocLinks: NodeListOf<Element>): TOCHierarchy {
    const hierarchy = new TOCHierarchy()
    // Determine heading level for each link (based on HTML structure or data attribute)
    Array.from(tocLinks).forEach((link : Element) => {
      // Try to determine level from data attribute first
      let level = 0 ; //parseInt(link.dataset.level) || 0;

      // If no data attribute, try to infer from parent elements or CSS classes
      if (!level) {
        // Option 1: Check if the link is in a nested list
        let nestingLevel = 0;
        let parent = link.parentElement;
        while (parent) {
          if (parent.tagName === 'UL' || parent.tagName === 'OL') {
            nestingLevel++;
          }
          parent = parent.parentElement;
        }
        level = nestingLevel || 1;

        // Option 2: Check for classes like 'toc-h2', 'toc-h3' etc.
        Array.from(link.classList).forEach(cls => {
          if (cls.match(/toc-h[1-6]/)) {
            level = parseInt(cls.substring(5));
          }
        });

        // Option 3: Check parent li for classes
        if (link.parentElement && link.parentElement.tagName === 'LI') {
          Array.from(link.parentElement.classList).forEach(cls => {
            if (cls.match(/level-[1-6]/)) {
              level = parseInt(cls.substring(6));
            }
          });
        }
      }

      // Store the level in our map
      hierarchy.levelMap.set(link, level || 1);
    });

    // Build parent-child relationships based on DOM structure and heading levels
    Array.from(tocLinks).forEach(link => {
      const linkLevel = hierarchy.levelMap.get(link);

      // Find the closest previous link with a higher level (parent)
      let parent = null;
      let previousSibling = this.getPreviousTocLink(link);

      while (previousSibling && !parent) {
        const siblingLevel = hierarchy.levelMap.get(previousSibling);

        if (siblingLevel && siblingLevel < linkLevel) {
          parent = previousSibling;
        } else {
          previousSibling = this.getPreviousTocLink(previousSibling);
        }
      }

      if (parent) {
        hierarchy.parentMap.set(link, parent);
      }
    });

    // Build ancestor and descendant maps
    Array.from(tocLinks).forEach(link => {
      // Build ancestors list for this link
      const ancestors = [];
      let current = link;

      while (hierarchy.parentMap.has(current)) {
        const parent = hierarchy.parentMap.get(current);
        ancestors.push(parent);
        current = parent;
      }

      hierarchy.ancestorMap.set(link, ancestors);

      // Add this link to the descendants list of all its ancestors
      ancestors.forEach(ancestor => {
        if (!hierarchy.descendantMap.has(ancestor)) {
          hierarchy.descendantMap.set(ancestor, []);
        }
        hierarchy.descendantMap.get(ancestor).push(link);
      });
    });

    return hierarchy;
  }

  // Helper function to get the previous TOC link in the DOM
  getPreviousTocLink(link: Element) {
    const config = this.config;
    let previous = link.previousElementSibling;

    // If no previous sibling, go up to parent and its previous sibling
    if (!previous) {
      const parentLi = link.closest('li');
      if (parentLi) {
        const parentUl = parentLi.parentElement;
        if (parentUl) {
          const previousLi = parentUl.previousElementSibling;
          if (previousLi) {
            // Find the last link in this list item
            const links = previousLi.querySelectorAll(config.tocLinkSelector);
            return links[links.length - 1] || null;
          }
        }
      }
      return null;
    }

    // If previous element is a list item, find its link
    if (previous.tagName === 'LI') {
      const links = previous.querySelectorAll(config.tocLinkSelector);
      return links[0] || null;
    }

    // If previous element is a link, return it
    if (previous.matches(config.tocLinkSelector)) {
      return previous;
    }

    return null;
  }
}
