// src/components/AdminUsuarios.tsx
// Administración de cuentas (Fase 7): listado paginado con búsqueda server-side,
// cambio de rol y activar/desactivar con confirmación.
import { useEffect, useState } from "react";
import {
  actualizarUsuario,
  getUsuarios,
  type UsuarioAdmin,
} from "../services/usuariosService";

const LIMITE = 10;

const AdminUsuarios = () => {
  const [usuarios, setUsuarios] = useState<UsuarioAdmin[]>([]);
  const [busqueda, setBusqueda] = useState<string>("");
  const [pagina, setPagina] = useState<number>(1);
  const [total, setTotal] = useState<number>(0);
  const [cargando, setCargando] = useState<boolean>(true);
  const [error, setError] = useState<string>("");

  const cargar = () => {
    setCargando(true);
    getUsuarios(busqueda, pagina, LIMITE)
      .then((res) => {
        setUsuarios(res.data);
        setTotal(res.total);
        setError("");
      })
      .catch((err) => setError(err.message || "No se pudieron cargar los usuarios"))
      .finally(() => setCargando(false));
  };

  // Búsqueda con debounce; recarga también al cambiar de página
  useEffect(() => {
    const timer = setTimeout(cargar, 300);
    return () => clearTimeout(timer);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [busqueda, pagina]);

  const cambiar = (u: UsuarioAdmin, cambios: { rol?: "admin" | "cliente"; activo?: boolean }, texto: string) => {
    if (!window.confirm(texto)) return;
    actualizarUsuario(u.id, cambios)
      .then(cargar)
      .catch((err) => window.alert(err.message || "No se pudo actualizar el usuario"));
  };

  const paginas = Math.max(1, Math.ceil(total / LIMITE));

  return (
    <div>
      <div className="mb-6">
        <h1 className="text-3xl uppercase text-marino-700">Administración de Usuarios</h1>
        <p className="text-slate-500 mt-1">Cuentas registradas, roles y estado de acceso.</p>
      </div>

      <div className="mb-4">
        <input
          type="search"
          value={busqueda}
          onChange={(e) => {
            setBusqueda(e.target.value);
            setPagina(1);
          }}
          placeholder="Buscar por nombre o correo..."
          className="w-full sm:w-96 px-4 py-2 border border-slate-300 focus:ring-2 focus:ring-indigo-600 focus:border-indigo-600 outline-none"
        />
      </div>

      {error && (
        <div className="bg-coral-50 text-coral-700 p-3 text-sm mb-4 border border-coral-200 border-l-4 border-l-coral-500">
          {error}
        </div>
      )}

      <div className="bg-papel border border-linea overflow-hidden">
        {cargando ? (
          <p className="p-6 text-slate-500">Cargando usuarios...</p>
        ) : usuarios.length === 0 ? (
          <p className="p-6 text-slate-500">No hay usuarios que coincidan con la búsqueda.</p>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full tabular-nums text-sm">
              <thead className="bg-arena-100 text-marino-700 uppercase tracking-widest text-xs border-b border-marino-700">
                <tr>
                  <th className="text-left p-3">Nombre</th>
                  <th className="text-left p-3">Correo</th>
                  <th className="text-center p-3">Rol</th>
                  <th className="text-center p-3">Estado</th>
                  <th className="text-right p-3">Acciones</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-linea">
                {usuarios.map((u) => (
                  <tr key={u.id} className="hover:bg-arena-50">
                    <td className="p-3 font-medium text-slate-800">{u.nombre || "—"}</td>
                    <td className="p-3 text-slate-600">{u.email}</td>
                    <td className="p-3 text-center">
                      <span
                        className={`text-xs font-semibold uppercase tracking-wider px-2 py-1 ${
                          u.rol === "admin" ? "bg-marino-700 text-arena-200" : "bg-slate-100 text-slate-700"
                        }`}
                      >
                        {u.rol === "admin" ? "Admin" : "Cliente"}
                      </span>
                    </td>
                    <td className="p-3 text-center">
                      <span
                        className={`text-xs font-semibold uppercase tracking-wider px-2 py-1 ${
                          u.activo ? "bg-turquesa-100 text-turquesa-800" : "bg-coral-100 text-coral-700"
                        }`}
                      >
                        {u.activo ? "Activo" : "Inactivo"}
                      </span>
                    </td>
                    <td className="p-3 text-right whitespace-nowrap">
                      <button
                        onClick={() =>
                          cambiar(
                            u,
                            { rol: u.rol === "admin" ? "cliente" : "admin" },
                            `¿Cambiar a ${u.nombre || u.email} al rol ${u.rol === "admin" ? "cliente" : "admin"}?`
                          )
                        }
                        className="text-indigo-600 hover:underline font-medium px-2 py-2 min-h-11"
                      >
                        {u.rol === "admin" ? "Hacer cliente" : "Hacer admin"}
                      </button>
                      <button
                        onClick={() =>
                          cambiar(
                            u,
                            { activo: !u.activo },
                            `¿Desea ${u.activo ? "desactivar" : "activar"} la cuenta de ${u.email}?`
                          )
                        }
                        className={`font-medium px-2 py-2 min-h-11 hover:underline ${
                          u.activo ? "text-coral-600" : "text-turquesa-700"
                        }`}
                      >
                        {u.activo ? "Desactivar" : "Activar"}
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

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
            Página {pagina} de {paginas} · {total} usuarios
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
    </div>
  );
};

export default AdminUsuarios;
