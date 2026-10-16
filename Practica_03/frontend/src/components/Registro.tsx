// src/components/Registro.tsx
// Alta pública de cuentas de cliente (Fase 1 del back-office).
// Si la URL trae ?ref=CODIGO, el campo viene prellenado y la cuenta nueva
// queda colgada del patrocinador en la red multinivel.
import { useState, type FormEvent } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import { useAuth, type Rol } from "../context/AuthContext";
import { apiFetch } from "../services/http";

const Registro = () => {
  const [nombre, setNombre] = useState<string>("");
  const [email, setEmail] = useState<string>("");
  const [password, setPassword] = useState<string>("");
  const [error, setError] = useState<string>("");
  const [loading, setLoading] = useState<boolean>(false);

  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const { login } = useAuth();

  // Prefill del código de invitación desde la URL (/registro?ref=LUIS).
  const refInicial = (searchParams.get("ref") ?? "").toUpperCase();
  const [ref, setRef] = useState<string>(refInicial);

  const handleSubmit = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setLoading(true);

    apiFetch<{ token?: string; message: string }>("/api/register", {
      method: "POST",
      body: JSON.stringify({
        nombre: nombre.trim(),
        email: email.trim(),
        password,
        ref: ref.trim().toUpperCase(),
      }),
    })
      .then(() => {
        // Tras registrar, iniciamos sesión automáticamente con las mismas credenciales.
        return apiFetch<{ token: string; email: string; rol: string }>("/api/login", {
          method: "POST",
          body: JSON.stringify({ email: email.trim(), password }),
        });
      })
      .then((sesion) => {
        setError("");
        const rol: Rol = sesion.rol === "admin" ? "admin" : "cliente";
        login({ email: sesion.email, rol }, sesion.token);
        navigate(rol === "admin" ? "/" : "/tienda");
      })
      .catch((err) => {
        setError(err.message || "No se pudo crear la cuenta");
      })
      .finally(() => {
        setLoading(false);
      });
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-slate-100 px-4 py-8">
      <div className="max-w-md w-full bg-white rounded-xl shadow-md p-8 border border-slate-200">
        <div className="text-center mb-8">
          <h2 className="text-3xl font-bold text-slate-900">Crear cuenta</h2>
          <p className="text-slate-500 mt-2">
            Únete a MultiCatálogo y empieza a comprar
          </p>
        </div>

        {error && (
          <div className="bg-red-50 text-red-600 p-3 rounded-lg text-sm mb-6 text-center border border-red-200">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-5">
          <div>
            <label className="block text-sm font-medium text-slate-700 mb-2">
              Nombre completo
            </label>
            <input
              type="text"
              value={nombre}
              onChange={(e) => setNombre(e.target.value)}
              className="w-full px-4 py-3 rounded-lg border border-slate-300 focus:ring-2 focus:ring-indigo-600 focus:border-indigo-600 outline-none transition"
              placeholder="Ana García"
              required
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-700 mb-2">
              Correo Electrónico
            </label>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="w-full px-4 py-3 rounded-lg border border-slate-300 focus:ring-2 focus:ring-indigo-600 focus:border-indigo-600 outline-none transition"
              placeholder="nombre@upse.edu.ec"
              required
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-700 mb-2">
              Contraseña
            </label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="w-full px-4 py-3 rounded-lg border border-slate-300 focus:ring-2 focus:ring-indigo-600 focus:border-indigo-600 outline-none transition"
              placeholder="••••••"
              minLength={6}
              required
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-700 mb-2">
              Código de referido <span className="text-slate-400 font-normal">(opcional)</span>
            </label>
            <input
              type="text"
              value={ref}
              onChange={(e) => setRef(e.target.value.toUpperCase())}
              className="w-full px-4 py-3 rounded-lg border border-slate-300 focus:ring-2 focus:ring-indigo-600 focus:border-indigo-600 outline-none transition font-mono uppercase"
              placeholder="LUIS"
            />
            {refInicial && (
              <p className="text-xs text-indigo-600 mt-1">
                Invitado por el código {refInicial}
              </p>
            )}
          </div>

          <button
            type="submit"
            disabled={loading}
            className="w-full bg-indigo-600 text-white font-bold py-3 rounded-lg hover:bg-indigo-700 transition disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {loading ? "Creando cuenta..." : "Crear Cuenta"}
          </button>
        </form>

        <p className="text-center text-sm text-slate-500 mt-6">
          ¿Ya tienes cuenta?{" "}
          <Link to="/login" className="text-indigo-600 font-semibold hover:underline">
            Inicia sesión
          </Link>
        </p>
      </div>
    </div>
  );
};

export default Registro;
