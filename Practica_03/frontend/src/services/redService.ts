// src/services/redService.ts
// Capa de servicio para la red multinivel (MLM) protegida por JWT.
import type { Referido } from "../data/red";
import { apiFetch } from "./http";

// Respuesta de /api/red/mia: el subárbol propio y el código de invitación.
export interface RedMia {
  red: Referido | null;
  codigo: string;
  referido_id: number;
}

export const getRed = async (): Promise<Referido> => {
  return apiFetch<Referido>("/api/red");
};

// GET /api/red/mia — subárbol que cuelga del nodo del usuario del token.
export const getRedMia = async (): Promise<RedMia> => {
  return apiFetch<RedMia>("/api/red/mia");
};
