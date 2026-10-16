// src/services/redService.ts
// Capa de servicio para la red multinivel (MLM) protegida por JWT.
import type { Referido } from "../data/red";
import { apiFetch } from "./http";

export const getRed = async (): Promise<Referido> => {
  return apiFetch<Referido>("/api/red");
};
