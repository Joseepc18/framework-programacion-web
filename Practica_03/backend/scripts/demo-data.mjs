// Carga datos de demostración recorriendo la API real (registro con código de
// referido, compras, cambios de estado y liquidación), de modo que stock,
// comisiones y contadores queden tal como los dejaría el uso normal.
//
// Uso:   node scripts/demo-data.mjs
// Env:   API_URL (por defecto http://localhost:3001)
//        ADMIN_EMAIL / ADMIN_PASSWORD (por defecto la cuenta semilla)
//        DEMO_PASSWORD (contraseña de las cuentas de demostración)
//
// Es re-ejecutable: las cuentas que ya existen se reutilizan y los pedidos
// solo se crean para quien todavía no tiene ninguno.

const API = (process.env.API_URL ?? 'http://localhost:3001').replace(/\/$/, '');
const ADMIN = {
  email: process.env.ADMIN_EMAIL ?? 'admin@upse.edu.ec',
  password: process.env.ADMIN_PASSWORD ?? '123456',
};
const DEMO_PASSWORD = process.env.DEMO_PASSWORD ?? 'Demo2026';

// Cadena de patrocinio: Luis (cuenta demo) → Valentina → Mateo → Camila.
// Así una compra de Camila reparte comisión a tres niveles de la red.
const PERSONAS = [
  { clave: 'valentina', nombre: 'Valentina Cedeño', email: 'valentina.cedeno@example.com', ref: 'LUIS', ciudad: 'Santa Elena' },
  { clave: 'mateo', nombre: 'Mateo Alvarado', email: 'mateo.alvarado@example.com', ref: 'valentina', ciudad: 'La Libertad' },
  { clave: 'camila', nombre: 'Camila Zambrano', email: 'camila.zambrano@example.com', ref: 'mateo', ciudad: 'Salinas' },
  { clave: 'andres', nombre: 'Andrés Villón', email: 'andres.villon@example.com', ref: 'ANA', ciudad: 'Guayaquil' },
  { clave: 'daniela', nombre: 'Daniela Reyes', email: 'daniela.reyes@example.com', ref: 'TU', ciudad: 'Quito' },
];

// Pedidos de la demo: quién compra, qué lleva y en qué estado termina.
// Cubre los cinco estados del ciclo de vida y las tres profundidades de comisión.
const PEDIDOS = [
  { quien: 'camila', items: [[8, 1], [6, 1]], estado: 'entregado' },
  { quien: 'camila', items: [[1, 2]], estado: 'entregado' },
  { quien: 'mateo', items: [[5, 1], [2, 1]], estado: 'enviado' },
  { quien: 'valentina', items: [[4, 2], [3, 1]], estado: 'preparando' },
  { quien: 'andres', items: [[1, 1], [7, 1]], estado: 'pendiente' },
  { quien: 'daniela', items: [[2, 1]], estado: 'cancelado' },
  { quien: 'luis', items: [[8, 1], [4, 1]], estado: 'entregado' },
  { quien: 'camila', items: [[3, 2], [6, 1]], estado: 'pendiente' },
  { quien: 'valentina', items: [[5, 1]], estado: 'entregado' },
  { quien: 'andres', items: [[8, 1]], estado: 'entregado' },
];

// Camino más corto en el ciclo de vida hasta cada estado final.
const RUTA = {
  pendiente: [],
  preparando: ['preparando'],
  enviado: ['preparando', 'enviado'],
  entregado: ['preparando', 'enviado', 'entregado'],
  cancelado: ['cancelado'],
};

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
  const cuentas = {
    luis: { nombre: 'Cliente Demo', email: 'cliente@upse.edu.ec', password: '123456', ciudad: 'Santa Elena' },
  };
  const codigos = { LUIS: 'LUIS', ANA: 'ANA', TU: 'TU' };

  // 1) Alta de cuentas con su patrocinador.
  for (const p of PERSONAS) {
    const ref = codigos[p.ref] ?? p.ref;
    const { status } = await llamar('POST', '/api/register', {
      cuerpo: { nombre: p.nombre, email: p.email, password: DEMO_PASSWORD, ref },
      aceptar: [409],
    });
    cuentas[p.clave] = { ...p, password: DEMO_PASSWORD };
    const token = await iniciarSesion(p.email, DEMO_PASSWORD);
    cuentas[p.clave].token = token;
    const { datos: mia } = await llamar('GET', '/api/red/mia', { token });
    codigos[p.clave] = mia.codigo;
    console.log(`${status === 201 ? 'creada ' : 'existía'} ${p.nombre} (código ${mia.codigo}, patrocinador ${ref})`);
  }
  cuentas.luis.token = await iniciarSesion(cuentas.luis.email, cuentas.luis.password);

  // 2) Pedidos: solo para quien todavía no tiene ninguno.
  const yaTienenPedidos = new Set();
  for (const [clave, cuenta] of Object.entries(cuentas)) {
    const { datos } = await llamar('GET', '/api/pedidos/mios?limit=1', { token: cuenta.token });
    if (datos.total > 0) yaTienenPedidos.add(clave);
  }

  const direccion = { Luis: 'Av. 9 de Octubre y Malecón' };
  let creados = 0;
  for (const pedido of PEDIDOS) {
    if (yaTienenPedidos.has(pedido.quien)) continue;
    const c = cuentas[pedido.quien];
    const { datos: nuevo } = await llamar('POST', '/api/checkout', {
      token: c.token,
      cuerpo: {
        nombre: c.nombre,
        email: c.email,
        direccion: direccion[c.nombre] ?? 'Calle Principal s/n',
        ciudad: c.ciudad,
        telefono: '0990000000',
        metodoPago: creados % 2 === 0 ? 'transferencia' : 'tarjeta',
        items: pedido.items.map(([producto_id, cantidad]) => ({ producto_id, cantidad })),
      },
    });
    for (const estado of RUTA[pedido.estado]) {
      await llamar('PATCH', `/api/pedidos/${nuevo.pedido_id}/estado`, { token: admin, cuerpo: { estado } });
    }
    creados++;
    // El número de pedido se arma con la hora al segundo: esperamos para que no se repita.
    await new Promise((listo) => setTimeout(listo, 1100));
    console.log(`pedido ${nuevo.numero} · ${c.nombre} · $${nuevo.total.toFixed(2)} → ${pedido.estado}`);
  }

  // 3) Una liquidación pagada a Valentina; el resto de saldos queda pendiente
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

  console.log(`\nListo: ${creados} pedidos nuevos. Contraseña de las cuentas demo: ${DEMO_PASSWORD}`);
}

main().catch((error) => {
  console.error(error.message);
  process.exit(1);
});
