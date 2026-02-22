// Path-based router for the Haira UI SPA.
// Routes: / (home), /workbench/<path>, /workflows, /agents, /observe, /settings, /deployments

export type Route =
  | { page: "home" }
  | { page: "workbench"; path: string }
  | { page: "workflows" }
  | { page: "agents" }
  | { page: "observe" }
  | { page: "settings" }
  | { page: "deployments" };

export function parseRoute(pathname: string): Route {
  const p = pathname || "/";

  if (p === "/") return { page: "home" };
  if (p === "/workflows") return { page: "workflows" };
  if (p === "/agents") return { page: "agents" };
  if (p === "/observe") return { page: "observe" };
  if (p === "/settings") return { page: "settings" };
  if (p === "/deployments") return { page: "deployments" };
  if (p.startsWith("/workbench/")) {
    return { page: "workbench", path: p.slice("/workbench".length) };
  }

  return { page: "home" };
}

export function navigate(route: Route) {
  let path: string;
  switch (route.page) {
    case "home":
      path = "/";
      break;
    case "workbench":
      path = `/workbench${route.path}`;
      break;
    default:
      path = `/${route.page}`;
      break;
  }
  history.pushState({}, "", path);
  // Dispatch popstate so all listeners update
  window.dispatchEvent(new PopStateEvent("popstate"));
}

export function currentRoute(): Route {
  return parseRoute(window.location.pathname);
}
