// src/services/checkoutService.ts
// Capa de servicios del checkout.
// Envía el pedido a POST /api/checkout (JWT automático): el backend valida stock, recalcula precios
// y guarda pedido, ítems y comisiones en una única transacción ACID.
import { apiFetch } from "./http";


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

// Respuesta cruda de POST /api/checkout.
interface RespuestaCheckout {
  pedido_id: number;
  numero: string;
  total: number;
  message: string;
}

// El backend toma la identidad del token y recalcula los precios: solo viajan producto_id y cantidad.
export const procesarCheckout = async (pedido: PedidoRequest): Promise<PedidoConfirmado> => {
  const res = await apiFetch<RespuestaCheckout>("/api/checkout", {
    method: "POST",
    body: JSON.stringify({
      ...pedido.cliente,
      metodoPago: pedido.metodoPago,
      items: pedido.items.map((i) => ({ producto_id: i.id, cantidad: i.cantidad })),
    }),
  });

  return {
    id: res.pedido_id,
    numero: res.numero,
    estado: "pendiente",
    total: res.total,
    fecha: new Date().toLocaleString("es-EC"),
    metodoPago: pedido.metodoPago,
    cliente: pedido.cliente,
    items: pedido.items,
  };
};
