// src/services/comisionesService.ts
// Capa de servicio del ledger de comisiones (Fase 4 del back-office).
import { apiFetch } from "./http";

// Estado posible de una fila del ledger.
export type EstadoComision = "pendiente" | "aprobada" | "pagada" | "revertida";

// Fila del ledger: qué pedido generó la comisión y cuánto pagó.
export interface Comision {
  id: number;
  pedido_id: number;
  pedido_numero: string;
  nivel: number;
  porcentaje: number;
  monto: number;
  estado: EstadoComision;
  creado_en: string;
}

// Respuesta estándar del ledger con los totales por estado.
export interface LedgerComisiones {
  data: Comision[];
  page: number;
  limit: number;
  total: number;
  pendiente: number;
  pagado: number;
}

// GET /api/comisiones/mias — ledger propio del usuario del token (paginado).
export const getMisComisiones = async (
  page = 1,
  limit = 10
): Promise<LedgerComisiones> => {
  return apiFetch<LedgerComisiones>(
    `/api/comisiones/mias?page=${page}&limit=${limit}`
  );
};
