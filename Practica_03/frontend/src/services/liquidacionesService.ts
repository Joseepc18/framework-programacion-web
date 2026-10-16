// src/services/liquidacionesService.ts
// Capa de servicio de liquidaciones (Fase 5 del back-office).
import { apiFetch } from "./http";

// Pago agrupado a un distribuidor.
export interface Liquidacion {
  id: number;
  referido_id: number;
  nombre: string;
  total: number;
  estado: "pendiente" | "pagada";
  creado_en: string;
  pagada_en?: string;
  cantidad_comisiones: number;
}

// Deuda pendiente de un distribuidor (cola de trabajo del admin).
export interface Deuda {
  referido_id: number;
  nombre: string;
  email: string;
  pendiente: number;
  cantidad: number;
}

// GET /api/liquidaciones/deudas — saldo pendiente por distribuidor (admin).
export const getDeudas = async (): Promise<Deuda[]> => {
  return apiFetch<Deuda[]>("/api/liquidaciones/deudas");
};

// GET /api/liquidaciones — histórico global de pagos (admin).
export const getLiquidaciones = async (
  page = 1,
  limit = 10
): Promise<{ data: Liquidacion[]; page: number; limit: number; total: number }> => {
  return apiFetch(`/api/liquidaciones?page=${page}&limit=${limit}`);
};

// POST /api/liquidaciones — liquida TODO el saldo de un distribuidor (admin).
export const liquidarDistribuidor = async (
  referido_id: number
): Promise<Liquidacion> => {
  return apiFetch<Liquidacion>("/api/liquidaciones", {
    method: "POST",
    body: JSON.stringify({ referido_id }),
  });
};

// GET /api/liquidaciones/mias — cobros recibidos por el usuario del token.
export const getMisLiquidaciones = async (
  page = 1,
  limit = 10
): Promise<{ data: Liquidacion[]; page: number; limit: number; total: number }> => {
  return apiFetch(`/api/liquidaciones/mias?page=${page}&limit=${limit}`);
};
