// Carga datos de demostración recorriendo la API real (altas con código de
// referido, catálogo, compras, cambios de estado y liquidación), de modo que
// stock, comisiones y contadores queden como los dejaría el uso normal.
//
// Uso:   node scripts/demo-data.mjs
// Env:   API_URL (por defecto http://localhost:3001)
//        ADMIN_EMAIL / ADMIN_PASSWORD (por defecto la cuenta semilla)
//
// Cada cuenta recibe su propia contraseña aleatoria, que se guarda en
// scripts/cuentas-demo.local (ignorado por git) para poder volver a ejecutarlo.
// Es re-ejecutable: lo que ya existe se reutiliza y los pedidos solo se crean
// para quien todavía no tiene ninguno. Las fechas se reparten después con
// scripts/demo-fechas.sql (la API siempre fecha los pedidos "ahora").

import { randomInt } from 'node:crypto';
import { existsSync, readFileSync, writeFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const API = (process.env.API_URL ?? 'http://localhost:3001').replace(/\/$/, '');
const ADMIN = {
  email: process.env.ADMIN_EMAIL ?? 'admin@upse.edu.ec',
  password: process.env.ADMIN_PASSWORD ?? '123456',
};
const ARCHIVO_CLAVES = join(dirname(fileURLToPath(import.meta.url)), 'cuentas-demo.local');

// Cuenta semilla del proyecto: es Luis Poveda dentro de la red.
const LUIS = { nombre: 'Luis Poveda', email: 'cliente@upse.edu.ec', password: '123456' };

// `ref` es el código del patrocinador: uno de la semilla (ANA, MARTA…) o la clave de otra cuenta.
const CUENTAS = [
  { clave: 'valentina', nombre: 'Valentina Cedeño', email: 'vale.cedeno@hotmail.com', ref: 'LUIS', ciudad: 'Santa Elena', direccion: 'Barrio Rocafuerte, calle 12 y av. 5', telefono: '0987412365' },
  { clave: 'mateo', nombre: 'Mateo Alvarado', email: 'mateoalvarado.97@gmail.com', ref: 'valentina', ciudad: 'La Libertad', direccion: 'Cdla. Virgen del Carmen, mz. 8 villa 14', telefono: '0996238841' },
  { clave: 'camila', nombre: 'Camila Zambrano', email: 'camizambrano@outlook.es', ref: 'mateo', ciudad: 'Salinas', direccion: 'Malecón de Chipipe, edif. Brisas 3B', telefono: '0961175209' },
  { clave: 'andres', nombre: 'Andrés Villón', email: 'avillon@gmail.com', ref: 'ANA', ciudad: 'Guayaquil', direccion: 'Urdesa Central, Las Lomas 410', telefono: '0983307716' },
  { clave: 'daniela', nombre: 'Daniela Reyes', email: 'dani.reyes@upse.edu.ec', ref: 'TU', ciudad: 'Manta', direccion: 'Av. 24 y calle 13, Barbasquillo', telefono: '0998854120' },
  { clave: 'joaquin', nombre: 'Joaquín Tomalá', email: 'jtomala@yahoo.com', ref: 'MARTA', ciudad: 'Playas', direccion: 'General Villamil, calle Los Cocoteros s/n', telefono: '0969402287' },
  { clave: 'kevin', nombre: 'Kevin Panchana', email: 'kpanchana@gmail.com', ref: 'daniela', ciudad: 'Portoviejo', direccion: 'Cdla. Los Bosques, calle B 207', telefono: '0992261548' },
  { clave: 'lissette', nombre: 'Lissette Quimí', email: 'lissquimi@hotmail.com', ref: 'CARLOS', ciudad: 'Guayaquil', direccion: 'Samborondón, Villa Club etapa Valencia', telefono: '0984716033' },
  // Se registró por invitación pero nunca compró y el admin ya la dio de baja.
  { clave: 'rosa', nombre: 'Rosa Mero', email: 'rosamero.ventas@gmail.com', ref: 'valentina', ciudad: 'La Libertad', direccion: 'Av. 9 de Octubre y calle 14', telefono: '0958839102', desactivar: true },
];

// Productos nuevos que amplían el catálogo semilla (solo categorías que el filtro ya conoce).
const PRODUCTOS_NUEVOS = [
  { nombre: 'Agua Micelar Calmante', categoria: 'Tónico', precio: 19.5, stock: 40, descripcion: 'Limpia y desmaquilla sin enjuague. Con pepino y aloe, apta para piel sensible.' },
  { nombre: 'Mascarilla de Arcilla Verde', categoria: 'Mascarilla', precio: 24, stock: 28, descripcion: 'Arcilla purificante para piel mixta o grasa. Controla el brillo y afina los poros.' },
  { nombre: 'Crema Corporal de Cacao', categoria: 'Crema', precio: 21.9, stock: 6, descripcion: 'Manteca de cacao y vitamina E para piel seca. Absorción rápida, aroma suave.' },
  { nombre: 'Set de Viaje Hidratación', categoria: 'Kit', precio: 39, stock: 15, descripcion: 'Cuatro formatos de 30 ml para llevar la rutina básica en el equipaje de mano.' },
  { nombre: 'Serum Edición Limitada Vitamina C', categoria: 'Serum', precio: 62, stock: 0, activo: false, descripcion: 'Edición de temporada agotada. Retirado del catálogo hasta la próxima producción.' },
];

// Pedidos: quién compra, qué lleva (por nombre de producto), cómo paga y en qué estado termina.
const TARJETA = 'Tarjeta de crédito';
const TRANSFERENCIA = 'Transferencia bancaria';
const EFECTIVO = 'Efectivo contra entrega';
const PEDIDOS = [
  { quien: 'camila', pago: TARJETA, estado: 'entregado', items: [['Kit Rutina Completa', 1], ['Crema Contorno de Ojos', 1]] },
  { quien: 'andres', pago: TRANSFERENCIA, estado: 'entregado', items: [['Kit Rutina Completa', 1]] },
  { quien: 'joaquin', pago: EFECTIVO, estado: 'entregado', items: [['Crema Contorno de Ojos', 2], ['Tónico Exfoliante AHA-BHA', 1], ['Tónico Purificante', 1]] },
  { quien: 'camila', pago: TRANSFERENCIA, estado: 'entregado', items: [['Serum Revitalizante', 2]] },
  { quien: 'luis', pago: TARJETA, estado: 'entregado', items: [['Kit Rutina Completa', 1], ['Mascarilla Nocturna', 1]] },
  { quien: 'valentina', pago: EFECTIVO, estado: 'entregado', items: [['Serum Anti-Edad Retinol', 1]] },
  { quien: 'daniela', pago: EFECTIVO, estado: 'cancelado', items: [['Crema Hidratante Pro', 1]] },
  { quien: 'mateo', pago: TRANSFERENCIA, estado: 'entregado', items: [['Serum Anti-Edad Retinol', 1], ['Crema Hidratante Pro', 1]] },
  { quien: 'lissette', pago: TARJETA, estado: 'cancelado', items: [['Serum Anti-Edad Retinol', 2]] },
  { quien: 'luis', pago: TRANSFERENCIA, estado: 'enviado', items: [['Crema Hidratante Pro', 3]] },
  { quien: 'kevin', pago: EFECTIVO, estado: 'enviado', items: [['Agua Micelar Calmante', 2], ['Mascarilla de Arcilla Verde', 1]] },
  { quien: 'mateo', pago: TARJETA, estado: 'preparando', items: [['Set de Viaje Hidratación', 1]] },
  { quien: 'valentina', pago: TARJETA, estado: 'preparando', items: [['Mascarilla Nocturna', 2], ['Tónico Purificante', 1]] },
  { quien: 'andres', pago: TRANSFERENCIA, estado: 'pendiente', items: [['Serum Revitalizante', 1], ['Tónico Exfoliante AHA-BHA', 1]] },
  { quien: 'camila', pago: EFECTIVO, estado: 'pendiente', items: [['Tónico Purificante', 2], ['Crema Corporal de Cacao', 1]] },
];

// Camino más corto en el ciclo de vida hasta cada estado final.
const RUTA = {
  pendiente: [],
  preparando: ['preparando'],
  enviado: ['preparando', 'enviado'],
  entregado: ['preparando', 'enviado', 'entregado'],
  cancelado: ['cancelado'],
};

const PALABRAS = ['Ceibo', 'Marea', 'Tamarindo', 'Coral', 'Guayacán', 'Brisa', 'Pelícano', 'Algarrobo', 'Salitre', 'Manglar', 'Cometa', 'Faro'];
const SIMBOLOS = ['#', '!', '*', '+', '_'];

function claveNueva() {
  const palabra = PALABRAS[randomInt(PALABRAS.length)];
  return `${palabra}${SIMBOLOS[randomInt(SIMBOLOS.length)]}${randomInt(100, 1000)}${PALABRAS[randomInt(PALABRAS.length)].slice(0, 3)}`;
}

function cargarClaves() {
  return existsSync(ARCHIVO_CLAVES) ? JSON.parse(readFileSync(ARCHIVO_CLAVES, 'utf8')) : {};
}

async function llamar(metodo, ruta, { token, cuerpo, aceptar = [] } = {}) {
  const respuesta = await fetch(`${API}${ruta}`, {
    method: metodo,
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
    body: cuerpo ? JSON.stringify(cuerpo) : undefined,
  });
  const datos = await respuesta.json().catch(() => ({}));
  if (!respuesta.ok && !aceptar.includes(respuesta.status)) {
    throw new Error(`${metodo} ${ruta} → ${respuesta.status} ${datos.message ?? ''}`);
  }
  return { status: respuesta.status, datos };
}

async function iniciarSesion(email, password) {
  const { datos } = await llamar('POST', '/api/login', { cuerpo: { email, password } });
  return datos.token;
}

async function main() {
  const { datos: salud } = await llamar('GET', '/api/health');
  console.log(`API ${API}: ${salud.status}`);

  const admin = await iniciarSesion(ADMIN.email, ADMIN.password);

  // 1) Catálogo: productos nuevos (el agotado y retirado se crea inactivo).
  const { datos: catalogo } = await llamar('GET', '/api/admin/productos?limit=100', { token: admin });
  const existentes = new Set((catalogo.data ?? []).map((p) => p.nombre));
  for (const p of PRODUCTOS_NUEVOS) {
    if (existentes.has(p.nombre)) continue;
    const semilla = encodeURIComponent(p.nombre.toLowerCase().replace(/\s+/g, '-'));
    const { datos: creado } = await llamar('POST', '/api/productos', {
      token: admin,
      cuerpo: {
        nombre: p.nombre,
        descripcion: p.descripcion,
        precio: p.precio,
        categoria: p.categoria,
        stock: p.stock,
        img: `https://picsum.photos/seed/${semilla}/600`,
        galeria: [1, 2, 3].map((n) => `https://picsum.photos/seed/${semilla}${n}/600`),
      },
    });
    if (p.activo === false) {
      await llamar('DELETE', `/api/productos/${creado.id}`, { token: admin });
    }
    console.log(`producto ${p.nombre}${p.activo === false ? ' (retirado)' : ''}`);
  }
  const { datos: publico } = await llamar('GET', '/api/productos');
  const idDe = Object.fromEntries((publico.data ?? publico).map((p) => [p.nombre, p.id]));

  // 2) Cuentas con su patrocinador y su propia contraseña.
  const claves = cargarClaves();
  const cuentas = { luis: { ...LUIS, ciudad: 'Santa Elena', direccion: 'Av. Principal y calle 6', telefono: '0991120045' } };
  const codigos = { LUIS: 'LUIS', ANA: 'ANA', MARTA: 'MARTA', CARLOS: 'CARLOS', TU: 'TU' };
  for (const p of CUENTAS) {
    claves[p.email] ??= claveNueva();
    const ref = codigos[p.ref] ?? p.ref;
    const { status } = await llamar('POST', '/api/register', {
      cuerpo: { nombre: p.nombre, email: p.email, password: claves[p.email], ref },
      aceptar: [409],
    });
    cuentas[p.clave] = { ...p, password: claves[p.email] };
    if (p.desactivar) {
      const { datos } = await llamar('GET', `/api/usuarios?q=${encodeURIComponent(p.email)}`, { token: admin });
      const usuario = (datos.data ?? []).find((u) => u.email === p.email);
      if (usuario?.activo) {
        await llamar('PATCH', `/api/usuarios/${usuario.id}`, { token: admin, cuerpo: { activo: false } });
      }
      console.log(`${status === 201 ? 'creada ' : 'existía'} ${p.nombre} (dada de baja)`);
      continue;
    }
    const token = await iniciarSesion(p.email, claves[p.email]);
    cuentas[p.clave].token = token;
    const { datos: mia } = await llamar('GET', '/api/red/mia', { token });
    codigos[p.clave] = mia.codigo;
    console.log(`${status === 201 ? 'creada ' : 'existía'} ${p.nombre} (código ${mia.codigo}, patrocinador ${ref})`);
  }
  writeFileSync(ARCHIVO_CLAVES, JSON.stringify(claves, null, 2));
  cuentas.luis.token = await iniciarSesion(LUIS.email, LUIS.password);

  // 3) Pedidos: solo para quien todavía no tiene ninguno.
  const yaTienenPedidos = new Set();
  for (const [clave, cuenta] of Object.entries(cuentas)) {
    if (!cuenta.token) continue;
    const { datos } = await llamar('GET', '/api/pedidos/mios?limit=1', { token: cuenta.token });
    if (datos.total > 0) yaTienenPedidos.add(clave);
  }

  let creados = 0;
  for (const pedido of PEDIDOS) {
    if (yaTienenPedidos.has(pedido.quien)) continue;
    const c = cuentas[pedido.quien];
    const { datos: nuevo } = await llamar('POST', '/api/checkout', {
      token: c.token,
      cuerpo: {
        nombre: c.nombre,
        email: c.email,
        direccion: c.direccion,
        ciudad: c.ciudad,
        telefono: c.telefono,
        metodoPago: pedido.pago,
        items: pedido.items.map(([nombre, cantidad]) => ({ producto_id: idDe[nombre], cantidad })),
      },
    });
    for (const estado of RUTA[pedido.estado]) {
      await llamar('PATCH', `/api/pedidos/${nuevo.pedido_id}/estado`, { token: admin, cuerpo: { estado } });
    }
    creados++;
    console.log(`pedido ${nuevo.numero} · ${c.nombre} · $${nuevo.total.toFixed(2)} → ${pedido.estado}`);
  }

  // 4) Una liquidación pagada a Valentina; el resto de saldos queda pendiente
  //    para que el admin pueda pagarlos en vivo durante la demostración.
  const { datos: mia } = await llamar('GET', '/api/red/mia', { token: cuentas.valentina.token });
  const { datos: liquidaciones } = await llamar('GET', '/api/liquidaciones', { token: admin });
  const yaLiquidada = (liquidaciones.data ?? []).some((l) => l.referido_id === mia.referido_id);
  if (!yaLiquidada) {
    const { status, datos } = await llamar('POST', '/api/liquidaciones', {
      token: admin,
      cuerpo: { referido_id: mia.referido_id },
      aceptar: [409],
    });
    console.log(status === 201 ? `liquidación pagada a Valentina: $${datos.total}` : 'Valentina sin saldo pendiente');
  }

  console.log(`\nListo: ${creados} pedidos nuevos. Contraseñas de las cuentas en ${ARCHIVO_CLAVES}`);
}

main().catch((error) => {
  console.error(error.message);
  process.exit(1);
});
