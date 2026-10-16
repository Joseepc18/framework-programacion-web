// src/services/reportesService.ts
// Capa de servicio de reportes de negocio (Fase 6, solo admin).
import { apiFetch } from "./http";
import type { Deuda } from "./liquidacionesService";

export interface ResumenVentas {
  total_vendido: number;
  num_pedidos: number;
  ticket_promedio: number;
  comisiones_devengadas: number;
}

export interface TopProducto {
  producto_id: number;
  nombre: string;
  unidades: number;
  ingresos: number;
}

export interface TopReferido {
  referido_id: number;
  nombre: string;
  ventas_red: number;
  comisiones: number;
}

// Rango opcional YYYY-MM-DD (ambos extremos inclusivos).
export interface RangoReporte {
  desde?: string;
  hasta?: string;
}

const consulta = (rango: RangoReporte): string => {
  const params = new URLSearchParams();
  if (rango.desde) params.set("desde", rango.desde);
  if (rango.hasta) params.set("hasta", rango.hasta);
  const texto = params.toString();
  return texto ? `?${texto}` : "";
};

export const getResumenVentas = (rango: RangoReporte = {}): Promise<ResumenVentas> =>
  apiFetch<ResumenVentas>(`/api/reportes/resumen${consulta(rango)}`);

export const getTopProductos = (rango: RangoReporte = {}): Promise<TopProducto[]> =>
  apiFetch<TopProducto[]>(`/api/reportes/top-productos${consulta(rango)}`);

export const getTopReferidos = (rango: RangoReporte = {}): Promise<TopReferido[]> =>
  apiFetch<TopReferido[]>(`/api/reportes/top-referidos${consulta(rango)}`);

export const getComisionesPendientes = (): Promise<Deuda[]> =>
  apiFetch<Deuda[]>("/api/reportes/comisiones-pendientes");
