// Carga datos de demostración recorriendo la API real (altas con código de
// referido, catálogo, compras, cambios de estado y liquidación), de modo que
// stock, comisiones y contadores queden como los dejaría el uso normal.
//
// Uso:   node scripts/demo-data.mjs
// Env:   API_URL (por defecto http://localhost:3001)
//        ADMIN_EMAIL / ADMIN_PASSWORD (por defecto la cuenta semilla)
//        SOLO_CATALOGO=1 carga únicamente los productos y termina
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

// Productos nuevos que amplían el catálogo semilla. Las categorías del filtro salen de los
// productos reales, así que basta con crearlos para que aparezcan.
const PRODUCTOS_NUEVOS = [
  { nombre: 'Agua Micelar Calmante', categoria: 'Tónico', precio: 19.5, stock: 40, descripcion: 'Limpia y desmaquilla sin enjuague. Con pepino y aloe, apta para piel sensible.' },
  { nombre: 'Mascarilla de Arcilla Verde', categoria: 'Mascarilla', precio: 24, stock: 28, descripcion: 'Arcilla purificante para piel mixta o grasa. Controla el brillo y afina los poros.' },
  { nombre: 'Crema Corporal de Cacao', categoria: 'Crema', precio: 21.9, stock: 6, descripcion: 'Manteca de cacao y vitamina E para piel seca. Absorción rápida, aroma suave.' },
  { nombre: 'Set de Viaje Hidratación', categoria: 'Kit', precio: 39, stock: 15, descripcion: 'Cuatro formatos de 30 ml para llevar la rutina básica en el equipaje de mano.' },
  { nombre: 'Mochila Escolar Reforzada', categoria: 'Utiles Escolares', precio: 32.5, stock: 18, descripcion: 'Tres compartimentos, espalda acolchada y base impermeable. Cabe un portátil de 14 pulgadas.', fotos: [
      'https://upload.wikimedia.org/wikipedia/commons/thumb/2/2c/School_bag_backpack.jpg/960px-School_bag_backpack.jpg',
      'https://upload.wikimedia.org/wikipedia/commons/thumb/e/ea/Back_to_school_snacks_water_bottle_notebook_backpack_%2820503811429%29.jpg/960px-Back_to_school_snacks_water_bottle_notebook_backpack_%2820503811429%29.jpg',
      'https://upload.wikimedia.org/wikipedia/commons/9/91/Backpacks_for_Guatemala_BGMC_01.jpg',
    ] },
  { nombre: 'Set de Marcadores Pastel x12', categoria: 'Utiles Escolares', precio: 8.9, stock: 60, descripcion: 'Doce tonos suaves de punta doble, fina y gruesa. Tinta a base de agua, sin olor.', fotos: [
      'https://upload.wikimedia.org/wikipedia/commons/thumb/2/29/Colored_whiteboard_markers_in_case.jpg/960px-Colored_whiteboard_markers_in_case.jpg',
      'https://upload.wikimedia.org/wikipedia/commons/thumb/b/ba/Colored_markers_%2812258455276%29.jpg/960px-Colored_markers_%2812258455276%29.jpg',
      'https://upload.wikimedia.org/wikipedia/commons/thumb/c/c1/Highlighting_pens.jpg/960px-Highlighting_pens.jpg',
    ] },
  { nombre: 'Calculadora Científica 240 Funciones', categoria: 'Utiles Escolares', precio: 24.5, stock: 14, descripcion: 'Pantalla de dos líneas con fracciones, estadística y ecuaciones. Permitida en exámenes.', fotos: [
      'https://upload.wikimedia.org/wikipedia/commons/thumb/6/6a/Casio_fx-991MS.jpg/960px-Casio_fx-991MS.jpg',
      'https://upload.wikimedia.org/wikipedia/commons/thumb/8/8f/Sharp_Scientific_Calculator.jpg/960px-Sharp_Scientific_Calculator.jpg',
      'https://upload.wikimedia.org/wikipedia/commons/8/81/Casio_fx-991MS_scientific_calculator_1aa.jpg',
    ] },
  { nombre: 'Lámpara de Escritorio LED', categoria: 'Hogar', precio: 27.9, stock: 22, descripcion: 'Brazo flexible, tres temperaturas de luz y regulación táctil. Carga USB incluida.', fotos: [
      'https://upload.wikimedia.org/wikipedia/commons/7/71/Concise_bamboo_eye_protection_LED_desk_lamp.jpg',
      'https://upload.wikimedia.org/wikipedia/commons/thumb/2/2c/Battery_powered_LED_desk_lamp-7420.jpg/960px-Battery_powered_LED_desk_lamp-7420.jpg',
      'https://upload.wikimedia.org/wikipedia/commons/thumb/0/06/Dekala_Arches%E2%84%A2_Smart_Lamp.jpg/960px-Dekala_Arches%E2%84%A2_Smart_Lamp.jpg',
    ] },
  { nombre: 'Juego de Sábanas de Algodón', categoria: 'Hogar', precio: 45, stock: 10, descripcion: 'Cama de dos plazas, algodón peinado de 200 hilos. Incluye funda de almohada.', fotos: [
      'https://upload.wikimedia.org/wikipedia/commons/thumb/3/3d/Bettwaesche.jpg/960px-Bettwaesche.jpg',
      'https://upload.wikimedia.org/wikipedia/commons/thumb/d/d2/Silk_Bedsheet.jpg/960px-Silk_Bedsheet.jpg',
      'https://upload.wikimedia.org/wikipedia/commons/thumb/8/8c/Bed_in_Seattle_hotel.jpg/960px-Bed_in_Seattle_hotel.jpg',
    ] },
  { nombre: 'Termo de Acero Inoxidable 750 ml', categoria: 'Hogar', precio: 18.5, stock: 35, descripcion: 'Mantiene el frío 24 horas y el calor 12. Tapa hermética y asa de transporte.', fotos: [
      'https://upload.wikimedia.org/wikipedia/commons/thumb/d/dd/Thermos_flask_Termoverken.jpg/960px-Thermos_flask_Termoverken.jpg',
      'https://upload.wikimedia.org/wikipedia/commons/thumb/1/1b/%D0%A2%D0%B5%D1%80%D0%BC%D0%BE%D1%81_%281%29.jpg/960px-%D0%A2%D0%B5%D1%80%D0%BC%D0%BE%D1%81_%281%29.jpg',
      'https://upload.wikimedia.org/wikipedia/commons/thumb/d/d5/Vintage_Aladdin_Vacuum_Flas.jpg/960px-Vintage_Aladdin_Vacuum_Flas.jpg',
    ] },
  { nombre: 'Audífonos Inalámbricos', categoria: 'Tecnología', precio: 39.9, stock: 25, descripcion: 'Bluetooth 5.3, cancelación básica de ruido y hasta 30 horas con el estuche.', fotos: [
      'https://upload.wikimedia.org/wikipedia/commons/thumb/5/56/2023_S%C5%82uchawki_Sony_WI-XB400_%281%29.jpg/960px-2023_S%C5%82uchawki_Sony_WI-XB400_%281%29.jpg',
      'https://upload.wikimedia.org/wikipedia/commons/thumb/a/a2/Sony_S_Wireless_Headphones_%2847103208%29.jpeg/960px-Sony_S_Wireless_Headphones_%2847103208%29.jpeg',
      'https://upload.wikimedia.org/wikipedia/commons/thumb/3/32/2023_S%C5%82uchawki_Sony_WI-XB400_%283%29.jpg/960px-2023_S%C5%82uchawki_Sony_WI-XB400_%283%29.jpg',
    ] },
  { nombre: 'Cargador USB-C 30 W', categoria: 'Tecnología', precio: 16.5, stock: 40, descripcion: 'Carga rápida para celular y tableta. Compacto, con protección contra sobrecalentamiento.', fotos: [
      'https://upload.wikimedia.org/wikipedia/commons/thumb/4/4a/Silicon_vs_GaN_30W_USB-C_chargers.jpg/960px-Silicon_vs_GaN_30W_USB-C_chargers.jpg',
      'https://upload.wikimedia.org/wikipedia/commons/thumb/8/86/Baseus_CCCP20UE_20W_USB_charger-0.jpg/960px-Baseus_CCCP20UE_20W_USB_charger-0.jpg',
      'https://upload.wikimedia.org/wikipedia/commons/thumb/0/01/MagSafe_and_USB-C_Cable_Charger_for_iPhone.jpg/960px-MagSafe_and_USB-C_Cable_Charger_for_iPhone.jpg',
    ] },
  { nombre: 'Parlante Bluetooth Portátil', categoria: 'Tecnología', precio: 34, stock: 3, descripcion: 'Resistente al agua, 12 horas de batería y micrófono integrado. Quedan pocas unidades.', fotos: [
      'https://upload.wikimedia.org/wikipedia/commons/thumb/a/a4/JBL_Pulse_5_Portable_Bluetooth_Speaker.jpg/960px-JBL_Pulse_5_Portable_Bluetooth_Speaker.jpg',
      'https://upload.wikimedia.org/wikipedia/commons/thumb/9/9d/UE_Boom_2.jpg/960px-UE_Boom_2.jpg',
      'https://upload.wikimedia.org/wikipedia/commons/thumb/2/2e/QFX_R-40_Portable_Bluetooth_Speaker.jpg/960px-QFX_R-40_Portable_Bluetooth_Speaker.jpg',
    ] },
  { nombre: 'Bandas Elásticas de Resistencia x3', categoria: 'Deportes', precio: 12.9, stock: 30, descripcion: 'Tres niveles de tensión para entrenar en casa. Incluye bolsa de transporte.', fotos: [
      'https://upload.wikimedia.org/wikipedia/commons/thumb/8/80/Woman_exercising_with_resistance_band_in_a_gym_setting.jpg/960px-Woman_exercising_with_resistance_band_in_a_gym_setting.jpg',
      'https://upload.wikimedia.org/wikipedia/commons/thumb/c/cf/Strength_band.png/960px-Strength_band.png',
      'https://upload.wikimedia.org/wikipedia/commons/7/72/Back-flyes-with-resistance-bands-1.png',
    ] },
  { nombre: 'Colchoneta de Yoga Antideslizante', categoria: 'Deportes', precio: 22, stock: 17, descripcion: 'Espesor de 6 mm, superficie texturizada y correa para cargarla.', fotos: [
      'https://upload.wikimedia.org/wikipedia/commons/thumb/6/6d/Yoga_mat.jpg/960px-Yoga_mat.jpg',
      'https://upload.wikimedia.org/wikipedia/commons/thumb/7/76/Yoga_mats.png/960px-Yoga_mats.png',
      'https://upload.wikimedia.org/wikipedia/commons/thumb/6/6e/Cotton_Yoga_Mats.png/960px-Cotton_Yoga_Mats.png',
    ] },
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
        img: p.fotos?.[0] ?? `https://picsum.photos/seed/${semilla}/600`,
        galeria: p.fotos?.slice(1) ?? [1, 2, 3].map((n) => `https://picsum.photos/seed/${semilla}${n}/600`),
      },
    });
    if (p.activo === false) {
      await llamar('DELETE', `/api/productos/${creado.id}`, { token: admin });
    }
    console.log(`producto ${p.nombre}${p.activo === false ? ' (retirado)' : ''}`);
  }
  if (process.env.SOLO_CATALOGO) {
    console.log('catálogo listo');
    return;
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
