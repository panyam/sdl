
INSTRUCTIONS
============

### Stack Details

* All component typescript files will be in the "./web/views/components" folder

* All html templates will be in the "./web/views/templates" folder.

* Tailwind config exists in ./web/tailwind.config.js

* We will only use typescript, html and tailwind.   

* DO NOT USE REACT OR ANGULAR

* Templates will be stored in the ./web/views/templates folder

* Only generate code incrementally in components.

* Make sure any time you have a block DO NOT GENERATE the details. Only generate a place holder like this:

```
<div name="BlockName" {{ styles and classes for this block }} >
    {{ template "BlockName.html" }}
</div>
```

This will ensure "BlockName" component can be as a fully refactored and decoupled component.

You can assume the following:

The following template exists that allows inclusion of other template components/blocks from another file:

```
{{# include "Component1.html" #}}
{{# include "AnotherComponent.html" #}}

<!-- And other components -->

```

After a component is included it can be invoked within other components via (note it does not have the ".html" suffix):

```
{{ template "Component1" . }}
```

This allows us to reuse components where necessary.

Before a compoment can be included in a file it must be included in the top of the file outside any "define"
declaratives.

You can assume that I have an engine that will render template components for testing.

Use typescript instead of javascript.  You can assume a webpack builder already exists and it will generate files with the prefix "gen/<Component>.html" (in the templates) so you do not have to generate webpack configs for me.  

When generating typescript files put them in their own files in the "./web/views/components" folder.

### Project Conventions

* Templates will contain all the top and component level templates

* Our convention is for root pages to be named templates/XYZPage.html where XYZ is the name of the page, eg DetailPage,
  ListingPage etc

* Every templates/XYZPage.html will have a corresponding components/XYZPage.ts for the page's
  entrypoint html that will setup all handlers, and be kicked off on DOMContentLoaded event.   If the page is using any
  third party React components it may be instead called components/XYZPage.tsx (to allow for embedding tsx).

* Each entry level page components/XYZPage.ts (or .tsx) is compiled by our webpack config into
  templates/gen/XYZPage.html so that this generated bundle is included by templates/XYZPage.html

* We are using templates/TemplateRegistry.html to store a bunch of client side templates.  This file should always be
  included via "{{# include ... #}}" after the body in the html.
