// Simple hash-based router for the Haira Console SPA.
// Routes: #/ (home), #/workbench/<path>, #/workflows, #/agents, #/observe, #/settings

export type Route =
  | { page: "home" }
  | { page: "workbench"; path: string }
  | { page: "workflows" }
  | { page: "agents" }
  | { page: "observe" }
  | { page: "settings" }
  | { page: "deployments" };

export function parseRoute(hash: string): Route {
  const h = hash.replace(/^#/, "") || "/";

  if (h === "/" || h === "") return { page: "home" };
  if (h === "/workflows") return { page: "workflows" };
  if (h === "/agents") return { page: "agents" };
  if (h === "/observe") return { page: "observe" };
  if (h === "/settings") return { page: "settings" };
  if (h === "/deployments") return { page: "deployments" };
  if (h.startsWith("/workbench/")) {
    return { page: "workbench", path: h.slice("/workbench".length) };
  }

  return { page: "home" };
}

export function navigate(route: Route) {
  let hash: string;
  switch (route.page) {
    case "home":
      hash = "#/";
      break;
    case "workbench":
      hash = `#/workbench${route.path}`;
      break;
    case "workflows":
      hash = "#/workflows";
      break;
    case "agents":
      hash = "#/agents";
      break;
    case "observe":
      hash = "#/observe";
      break;
    case "settings":
      hash = "#/settings";
      break;
    case "deployments":
      hash = "#/deployments";
      break;
  }
  window.location.hash = hash;
}

export function currentRoute(): Route {
  return parseRoute(window.location.hash);
}
