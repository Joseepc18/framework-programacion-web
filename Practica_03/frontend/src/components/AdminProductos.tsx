// src/components/AdminProductos.tsx
// Administración del catálogo (Fase 2): tabla con búsqueda server-side (?q=),
// alta/edición en formulario y borrado lógico con confirmación.
import { useEffect, useState, type FormEvent } from "react";
import ImagenProducto from "./ImagenProducto";
import type { Producto } from "../data/productos";
import {
  getProductosAdmin,
  crearProducto,
  actualizarProducto,
  cambiarEstadoProducto,
} from "../services/adminProductosService";

// Estado vacío del formulario de alta/edición
const PRODUCTO_VACIO = {
  nombre: "",
  descripcion: "",
  precio: 0,
  categoria: "",
  img: "",
  galeria: [] as string[],
  stock: 50,
};

type FormProducto = typeof PRODUCTO_VACIO;

// Comprueba si la URL realmente carga como imagen y muestra una vista previa o un aviso.
// Es solo informativo: no impide guardar.
const VistaPreviaImagen = ({ url }: { url: string }) => {
  const [resultado, setResultado] = useState<{ url: string; ok: boolean } | null>(null);
  const direccion = url.trim();

  useEffect(() => {
    if (!direccion) return;
    const sonda = new Image();
    sonda.onload = () => setResultado({ url: direccion, ok: true });
    sonda.onerror = () => setResultado({ url: direccion, ok: false });
    const timer = setTimeout(() => {
      sonda.src = direccion;
    }, 400);
    return () => {
      clearTimeout(timer);
      sonda.onload = null;
      sonda.onerror = null;
    };
  }, [direccion]);

  if (!direccion) return null;
  if (resultado?.url !== direccion) {
    return <p className="mt-2 text-xs text-slate-500">Comprobando imagen...</p>;
  }
  if (!resultado.ok) {
    return (
      <p className="mt-2 border border-coral-200 border-l-4 border-l-coral-500 bg-coral-50 p-2 text-xs text-coral-700">
        Esta dirección no se puede mostrar como imagen. Usa el enlace directo al archivo (suele
        terminar en .jpg, .png o .webp), no la página donde se ve la foto.
      </p>
    );
  }
  return <img src={direccion} alt="Vista previa" className="mt-2 h-28 w-28 border border-linea object-cover" />;
};

const AdminProductos = () => {
  const [productos, setProductos] = useState<Producto[]>([]);
  const [busqueda, setBusqueda] = useState<string>("");
  const [cargando, setCargando] = useState<boolean>(true);
  const [error, setError] = useState<string>("");

  // Formulario (null = cerrado; con id = edición; sin id = alta)
  const [form, setForm] = useState<(FormProducto & { id?: number }) | null>(null);
  const [guardando, setGuardando] = useState<boolean>(false);
  const [errorForm, setErrorForm] = useState<string>("");

  const [pagina, setPagina] = useState<number>(1);
  const [total, setTotal] = useState<number>(0);
  const LIMITE = 10;

  const cargar = (q = busqueda, pag = pagina) => {
    setCargando(true);
    getProductosAdmin(q, pag, LIMITE)
      .then((res) => {
        setProductos(res.data);
        setTotal(res.total);
        setError("");
      })
      .catch((err) => {
        setError(err.message || "No se pudo cargar el catálogo administrativo");
      })
      .finally(() => setCargando(false));
  };

  // Búsqueda server-side con debounce para no saturar la API; recarga también al paginar
  useEffect(() => {
    const timer = setTimeout(() => {
      cargar(busqueda, pagina);
    }, 300);
    return () => clearTimeout(timer);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [busqueda, pagina]);

  const paginas = Math.max(1, Math.ceil(total / LIMITE));

  const abrirAlta = () => {
    setForm({ ...PRODUCTO_VACIO, galeria: [] });
    setErrorForm("");
  };

  const abrirEdicion = (p: Producto) => {
    setForm({
      id: p.id,
      nombre: p.nombre,
      descripcion: p.descripcion,
      precio: p.precio,
      categoria: p.categoria,
      img: p.img,
      galeria: p.galeria ?? [],
      stock: p.stock ?? 0,
    });
    setErrorForm("");
  };

  const guardar = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!form) return;
    setGuardando(true);
    setErrorForm("");

    const cuerpo = {
      nombre: form.nombre.trim(),
      descripcion: form.descripcion.trim(),
      precio: Number(form.precio),
      categoria: form.categoria.trim(),
      img: form.img.trim(),
      galeria: form.galeria,
      stock: Number(form.stock),
    };

    const operacion =
      form.id === undefined
        ? crearProducto(cuerpo)
        : actualizarProducto(form.id, cuerpo);

    operacion
      .then(() => {
        setForm(null);
        cargar(busqueda);
      })
      .catch((err) => {
        setErrorForm(err.message || "No se pudo guardar el producto");
      })
      .finally(() => setGuardando(false));
  };

  const alternarEstado = (p: Producto) => {
    const activoActual = p.activo ?? true;
    const accion = activoActual ? "desactivar" : "activar";
    if (!window.confirm(`¿Desea ${accion} el producto "${p.nombre}"?`)) return;

    cambiarEstadoProducto(p.id, activoActual)
      .then(() => cargar(busqueda))
      .catch((err) => {
        window.alert(err.message || "No se pudo cambiar el estado");
      });
  };

  const cambiarCampo = (campo: keyof FormProducto, valor: string | number | string[]) => {
    setForm((prev) => (prev ? { ...prev, [campo]: valor } : prev));
  };

  return (
    <div>
      <div className="flex flex-wrap items-center justify-between gap-3 mb-6">
        <div>
          <h1 className="text-3xl uppercase text-marino-700">Administración de Productos</h1>
          <p className="text-slate-500 mt-1">
            Alta, edición y borrado lógico del catálogo.
          </p>
        </div>
        <button
          onClick={abrirAlta}
          className="bg-marino-700 text-arena-200 font-semibold px-4 py-2 hover:bg-marino-800 transition"
        >
          + Nuevo producto
        </button>
      </div>

      {/* Búsqueda server-side */}
      <div className="mb-4">
        <input
          type="search"
          value={busqueda}
          onChange={(e) => {
            setBusqueda(e.target.value);
            setPagina(1);
          }}
          placeholder="Buscar por nombre, categoría o descripción..."
          className="w-full sm:w-96 px-4 py-2 border border-slate-300 focus:ring-2 focus:ring-indigo-600 focus:border-indigo-600 outline-none"
        />
      </div>

      {error && (
        <div className="bg-coral-50 text-coral-700 p-3 text-sm mb-4 border border-coral-200 border-l-4 border-l-coral-500">
          {error}
        </div>
      )}

      {/* Tarjetas en móvil, tabla en escritorio */}
      <div className="bg-papel border border-linea overflow-hidden">
        {cargando ? (
          <p className="p-6 text-slate-500">Cargando catálogo...</p>
        ) : productos.length === 0 ? (
          <p className="p-6 text-slate-500">No hay productos que coincidan con la búsqueda.</p>
        ) : (
          <>
            {/* Tabla (escritorio) */}
            <table className="hidden md:table w-full text-sm">
              <thead className="bg-arena-100 text-marino-700 uppercase tracking-widest text-xs border-b border-marino-700">
                <tr>
                  <th className="text-left p-3">Producto</th>
                  <th className="text-left p-3">Categoría</th>
                  <th className="text-right p-3">Precio</th>
                  <th className="text-right p-3">Stock</th>
                  <th className="text-center p-3">Estado</th>
                  <th className="text-right p-3">Acciones</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-linea">
                {productos.map((p) => (
                  <tr key={p.id} className={p.activo === false ? "bg-slate-100 opacity-70" : "hover:bg-arena-50"}>
                    <td className="p-3">
                      <div className="flex items-center gap-3">
                        <ImagenProducto src={p.img} alt={p.nombre} className="w-10 h-10 object-cover" />
                        <span className="font-medium text-slate-800">{p.nombre}</span>
                      </div>
                    </td>
                    <td className="p-3 text-slate-600">{p.categoria}</td>
                    <td className="p-3 text-right">${p.precio.toFixed(2)}</td>
                    <td className="p-3 text-right">{p.stock ?? 0}</td>
                    <td className="p-3 text-center">
                      <span
                        className={`text-xs font-semibold uppercase tracking-wider px-2 py-1 ${
                          p.activo === false
                            ? "bg-coral-100 text-coral-700"
                            : "bg-turquesa-100 text-turquesa-800"
                        }`}
                      >
                        {p.activo === false ? "Inactivo" : "Activo"}
                      </span>
                    </td>
                    <td className="p-3 text-right space-x-2 whitespace-nowrap">
                      <button
                        onClick={() => abrirEdicion(p)}
                        className="text-indigo-600 hover:underline font-medium"
                      >
                        Editar
                      </button>
                      <button
                        onClick={() => alternarEstado(p)}
                        className={p.activo === false ? "text-turquesa-700 hover:underline font-medium" : "text-coral-600 hover:underline font-medium"}
                      >
                        {p.activo === false ? "Activar" : "Desactivar"}
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>

            {/* Tarjetas (móvil) */}
            <ul className="md:hidden divide-y divide-linea">
              {productos.map((p) => (
                <li key={p.id} className="p-4">
                  <div className="flex items-center gap-3">
                    <ImagenProducto src={p.img} alt={p.nombre} className="w-12 h-12 object-cover" />
                    <div className="flex-1">
                      <p className="font-semibold text-slate-800">{p.nombre}</p>
                      <p className="text-sm text-slate-500">
                        {p.categoria} · ${p.precio.toFixed(2)} · Stock: {p.stock ?? 0}
                      </p>
                      <span
                        className={`inline-block mt-1 text-xs font-semibold uppercase tracking-wider px-2 py-0.5 ${
                          p.activo === false
                            ? "bg-coral-100 text-coral-700"
                            : "bg-turquesa-100 text-turquesa-800"
                        }`}
                      >
                        {p.activo === false ? "Inactivo" : "Activo"}
                      </span>
                    </div>
                  </div>
                  <div className="mt-3 flex gap-4">
                    <button
                      onClick={() => abrirEdicion(p)}
                      className="text-indigo-600 hover:underline font-medium text-sm"
                    >
                      Editar
                    </button>
                    <button
                      onClick={() => alternarEstado(p)}
                      className={p.activo === false ? "text-turquesa-700 hover:underline font-medium text-sm" : "text-coral-600 hover:underline font-medium text-sm"}
                    >
                      {p.activo === false ? "Activar" : "Desactivar"}
                    </button>
                  </div>
                </li>
              ))}
            </ul>
          </>
        )}
      </div>

      {/* Paginación */}
      {paginas > 1 && (
        <div className="flex items-center justify-between mt-4 text-sm">
          <button
            onClick={() => setPagina((p) => Math.max(1, p - 1))}
            disabled={pagina === 1}
            className="px-3 py-2 min-h-11 border border-marino-700 text-marino-700 font-medium disabled:opacity-40 hover:bg-marino-700 hover:text-arena-200"
          >
            Anterior
          </button>
          <span className="text-slate-500">
            Página {pagina} de {paginas} · {total} productos
          </span>
          <button
            onClick={() => setPagina((p) => Math.min(paginas, p + 1))}
            disabled={pagina === paginas}
            className="px-3 py-2 min-h-11 border border-marino-700 text-marino-700 font-medium disabled:opacity-40 hover:bg-marino-700 hover:text-arena-200"
          >
            Siguiente
          </button>
        </div>
      )}

      {/* ===== FORMULARIO (modal) ===== */}
      {form && (
        <div className="fixed inset-0 z-50 flex items-end sm:items-center justify-center bg-black/50 p-0 sm:p-4">
          <div className="bg-papel w-full sm:max-w-lg p-6 max-h-[90vh] overflow-y-auto">
            <h2 className="text-lg font-bold text-slate-800 mb-4">
              {form.id === undefined ? "Nuevo producto" : `Editar producto #${form.id}`}
            </h2>

            {errorForm && (
              <div className="bg-coral-50 text-coral-700 p-3 text-sm mb-4 border border-coral-200 border-l-4 border-l-coral-500">
                {errorForm}
              </div>
            )}

            <form onSubmit={guardar} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-slate-700 mb-1">Nombre</label>
                <input
                  type="text"
                  value={form.nombre}
                  onChange={(e) => cambiarCampo("nombre", e.target.value)}
                  className="w-full px-3 py-2 border border-slate-300 focus:ring-2 focus:ring-indigo-600 outline-none"
                  required
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-slate-700 mb-1">Descripción</label>
                <textarea
                  value={form.descripcion}
                  onChange={(e) => cambiarCampo("descripcion", e.target.value)}
                  rows={3}
                  className="w-full px-3 py-2 border border-slate-300 focus:ring-2 focus:ring-indigo-600 outline-none"
                />
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-sm font-medium text-slate-700 mb-1">Precio ($)</label>
                  <input
                    type="number"
                    step="0.01"
                    min="0.01"
                    value={form.precio}
                    onChange={(e) => cambiarCampo("precio", Number(e.target.value))}
                    className="w-full px-3 py-2 border border-slate-300 focus:ring-2 focus:ring-indigo-600 outline-none"
                    required
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-slate-700 mb-1">Stock</label>
                  <input
                    type="number"
                    min="0"
                    value={form.stock}
                    onChange={(e) => cambiarCampo("stock", Number(e.target.value))}
                    className="w-full px-3 py-2 border border-slate-300 focus:ring-2 focus:ring-indigo-600 outline-none"
                    required
                  />
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-slate-700 mb-1">Categoría</label>
                <input
                  type="text"
                  value={form.categoria}
                  onChange={(e) => cambiarCampo("categoria", e.target.value)}
                  placeholder="Ej.: Hogar, Tecnología, Útiles Escolares"
                  className="w-full px-3 py-2 border border-slate-300 focus:ring-2 focus:ring-indigo-600 outline-none"
                  required
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-slate-700 mb-1">URL de imagen</label>
                <input
                  type="url"
                  value={form.img}
                  onChange={(e) => cambiarCampo("img", e.target.value)}
                  placeholder="https://..."
                  className="w-full px-3 py-2 border border-slate-300 focus:ring-2 focus:ring-indigo-600 outline-none"
                  required
                />
                <VistaPreviaImagen url={form.img} />
              </div>

              <div className="flex gap-3 pt-2">
                <button
                  type="submit"
                  disabled={guardando}
                  className="flex-1 bg-marino-700 text-arena-200 font-semibold py-2 hover:bg-marino-800 transition disabled:opacity-50"
                >
                  {guardando ? "Guardando..." : "Guardar"}
                </button>
                <button
                  type="button"
                  onClick={() => setForm(null)}
                  className="px-4 py-2 border border-slate-300 text-slate-700 hover:bg-slate-50 transition"
                >
                  Cancelar
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};

export default AdminProductos;
