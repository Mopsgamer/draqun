import "./main.ts";
import { SwaggerUIBundle } from "swagger-ui-dist";
SwaggerUIBundle({
    url: "/static/assets/swagger.yaml",
    dom_id: "#swagger-ui",
});
