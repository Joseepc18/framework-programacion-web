// src/services/pedidosService.ts
// Capa de servicio del módulo de pedidos (Fase 3 del back-office).
// El historial propio y el detalle viajan con el token JWT del usuario.
import { apiFetch } from "./http";

// Estados posibles del ciclo de vida de un pedido.
export type EstadoPedido =
  | "pendiente"
  | "preparando"
  | "enviado"
  | "entregado"
  | "cancelado";

// Cabecera de un pedido tal como la devuelve la API.
export interface Pedido {
  id: number;
  numero: string;
  cliente_nombre: string;
  cliente_email: string;
  cliente_direccion: string;
  cliente_ciudad: string;
  cliente_telefono: string;
  metodo_pago: string;
  total: number;
  usuario_id?: number;
  estado: EstadoPedido;
  creado_en: string;
}

// Línea de pedido con el precio vigente al momento de comprar.
export interface ItemPedido {
  producto_id: number;
  nombre: string;
  cantidad: number;
  precio_unitario: number;
  subtotal: number;
}

// Cabecera + líneas de un pedido concreto.
export interface PedidoDetalle extends Pedido {
  items: ItemPedido[];
}

// Forma estándar de respuesta paginada de la API.
export interface PaginaPedidos {
  data: Pedido[];
  page: number;
  limit: number;
  total: number;
}

// GET /api/pedidos/mios — historial del usuario del token (paginado).
export const getMisPedidos = async (
  page = 1,
  limit = 10
): Promise<PaginaPedidos> => {
  return apiFetch<PaginaPedidos>(
    `/api/pedidos/mios?page=${page}&limit=${limit}`
  );
};

// GET /api/pedidos — back-office con filtro opcional por estado (solo admin).
export const getPedidosAdmin = async (
  estado?: string,
  page = 1,
  limit = 10
): Promise<PaginaPedidos> => {
  const params = new URLSearchParams({ page: String(page), limit: String(limit) });
  if (estado && estado.trim() !== "") params.set("estado", estado.trim());
  return apiFetch<PaginaPedidos>(`/api/pedidos?${params.toString()}`);
};

// GET /api/pedidos/:id — detalle con líneas (dueño o admin).
export const getPedidoDetalle = async (id: number): Promise<PedidoDetalle> => {
  return apiFetch<PedidoDetalle>(`/api/pedidos/${id}`);
};

// PATCH /api/pedidos/:id/estado — transición de ciclo de vida (solo admin).
export const cambiarEstadoPedido = async (
  id: number,
  estado: EstadoPedido
): Promise<Pedido> => {
  return apiFetch<Pedido>(`/api/pedidos/${id}/estado`, {
    method: "PATCH",
    body: JSON.stringify({ estado }),
  });
};
