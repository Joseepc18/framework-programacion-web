// src/services/checkoutService.ts
// Capa de servicios del checkout.
// Envía el pedido a POST /api/checkout, donde el backend lo guarda en una transacción ACID.

const API_URL = "http://localhost:3000/api";

// Datos del formulario de envío
export interface DatosCliente {
  nombre: string;
  email: string;
  direccion: string;
  ciudad: string;
  telefono: string;
}

// Un producto del carrito con su cantidad
export interface ItemPedido {
  id: number;
  nombre: string;
  cantidad: number;
  precio: number;
}

// Cuerpo que espera el backend
export interface PedidoRequest {
  cliente: DatosCliente;
  usuarioEmail: string;
  metodoPago: string;
  items: ItemPedido[];
}

// Pedido ya registrado que devuelve la API (lo usa Confirmacion.tsx)
export interface PedidoConfirmado {
  id: number;
  numero: string;
  estado: string;
  total: number;
  fecha: string;
  metodoPago: string;
  cliente: DatosCliente;
  items: ItemPedido[];
}

// POST /api/checkout (registra la compra y dispara las comisiones en segundo plano)
export const procesarCheckout = async (pedido: PedidoRequest): Promise<PedidoConfirmado> => {
  const response = await fetch(`${API_URL}/checkout`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(pedido),
  });

  // Si la transacción falló, el backend responde APIError con el motivo: lo mostramos tal cual.
  if (!response.ok) {
    const error = await response.json().catch(() => null);
    throw new Error(error?.message ?? "No se pudo procesar el pedido");
  }

  return response.json();
};
