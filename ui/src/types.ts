export interface ServiceConfig {
  meilisearch: boolean;
  trivy: boolean;
  openscap: boolean;
  dependencyTrack: boolean;
  jaeger: boolean;
}

export interface Config {
  port: number;
  bindAddress: string;
  exposePostgres: boolean;
  services: ServiceConfig;
}

export interface ServiceHealth {
  name: string;
  status: "healthy" | "unhealthy" | "starting" | "not_running";
  running: boolean;
  image?: string;
}

export interface HealthReport {
  overall: string;
  services: ServiceHealth[];
}

export interface SecretsInfo {
  isFirstRun: boolean;
  adminPassword: string;
}

export interface ExtensionInfo {
  extensionVersion: string;
  backendVersion: string;
  webVersion: string;
  port: number;
  bindAddress: string;
}

export const DEFAULT_CONFIG: Config = {
  port: 8080,
  bindAddress: "127.0.0.1",
  exposePostgres: false,
  services: {
    meilisearch: true,
    trivy: false,
    openscap: false,
    dependencyTrack: false,
    jaeger: false,
  },
};

export const SERVICE_INFO: Record<
  keyof ServiceConfig,
  { label: string; description: string; ram: string; warning?: string }
> = {
  meilisearch: {
    label: "Meilisearch",
    description: "Full-text search engine for artifacts and packages",
    ram: "~128 MB",
  },
  trivy: {
    label: "Trivy",
    description: "Vulnerability scanner for container images and artifacts",
    ram: "~512 MB",
    warning:
      "Trivy downloads vulnerability databases on first run, which may take several minutes.",
  },
  openscap: {
    label: "OpenSCAP",
    description: "Security compliance scanning and policy evaluation",
    ram: "~256 MB",
    warning: "Requires additional SCAP content profiles to be configured.",
  },
  dependencyTrack: {
    label: "Dependency-Track",
    description:
      "Software composition analysis for SBOM ingestion and license tracking",
    ram: "~1 GB",
    warning:
      "Dependency-Track requires significant memory and may slow down smaller machines.",
  },
  jaeger: {
    label: "Jaeger",
    description: "Distributed tracing for request flow visualization",
    ram: "~256 MB",
  },
};
