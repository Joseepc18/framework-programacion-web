// src/components/Login.tsx
import { useState, type FormEvent } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useAuth, type Rol } from "../context/AuthContext";
import { apiFetch } from "../services/http";
import InstalarApp from './InstalarApp';
import Logo from "./Logo";

const Login = () => {
  const [email, setEmail] = useState<string>("");
  const [password, setPassword] = useState<string>("");
  const [error, setError] = useState<string>("");
  const [loading, setLoading] = useState<boolean>(false);

  const navigate = useNavigate();
  const { login } = useAuth();

  const handleSubmit = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setLoading(true);

    apiFetch<{ token: string; email: string; rol: string }>("/api/login", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    })
      .then((data) => {
        setError("");
        const rol: Rol = data.rol === "admin" ? "admin" : "cliente";
        login({ email: data.email, rol }, data.token);

        // Redirigimos según el rol: admin al Dashboard, cliente a la Tienda
        navigate(rol === "admin" ? "/" : "/tienda");
      })
      .catch((err) => {
        setError(err.message || "Credenciales incorrectas");
      })
      .finally(() => {
        setLoading(false);
      });
  };

  const campo =
    "w-full border border-linea bg-white px-4 py-3 text-tinta placeholder:text-piedra outline-none transition focus:border-marino-600 focus:ring-1 focus:ring-marino-600";
  const etiqueta = "mb-2 block text-xs font-semibold uppercase tracking-widest text-slate-600";

  return (
    <div className="grid min-h-screen bg-fondo md:grid-cols-2">
      <aside className="hidden flex-col justify-between bg-marino-700 p-12 text-arena-200 md:flex">
        <Logo tono="claro" size={44} />
        <div>
          <h1 className="max-w-md font-display text-5xl leading-tight">
            Compra, vende y haz crecer tu red.
          </h1>
          <p className="mt-6 max-w-md text-marino-200">
            Cada venta de tu red te devuelve una comisión: 10 % del primer nivel,
            5 % del segundo y 2 % del tercero.
          </p>
        </div>
        <p className="text-xs uppercase tracking-widest text-marino-300">Jose Ponce Carlos</p>
      </aside>

      <main className="flex items-center justify-center px-6 py-12">
        <div className="w-full max-w-sm">
          <div className="mb-10 md:hidden">
            <Logo />
          </div>

          <h2 className="font-display text-4xl text-tinta">Ingresar</h2>
          <p className="mt-2 text-slate-600">Accede a tu cuenta para continuar.</p>

          {error && (
            <div className="mt-6 border-l-4 border-coral-500 bg-coral-50 p-3 text-sm text-coral-700">
              {error}
            </div>
          )}

          <form onSubmit={handleSubmit} className="mt-8 space-y-5">
            <div>
              <label className={etiqueta}>Correo electrónico</label>
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className={campo}
                placeholder="admin@upse.edu.ec"
                required
              />
            </div>

            <div>
              <label className={etiqueta}>Contraseña</label>
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className={campo}
                placeholder="••••••"
                required
              />
            </div>

            <button
              type="submit"
              disabled={loading}
              className="w-full bg-marino-700 py-3 font-semibold text-arena-200 transition hover:bg-marino-800 disabled:cursor-not-allowed disabled:opacity-50"
            >
              {loading ? "Validando..." : "Iniciar sesión"}
            </button>
          </form>

          <InstalarApp ubicacion="login" />

          <p className="mt-6 text-sm text-slate-600">
            ¿Aún no tienes cuenta?{" "}
            <Link to="/registro" className="font-semibold text-coral-600 underline underline-offset-4 hover:text-coral-700">
              Crear cuenta
            </Link>
          </p>

          <dl className="mt-10 border-t border-linea pt-4 text-xs text-slate-600">
            <dt className="font-semibold uppercase tracking-widest text-slate-700">Cuentas de prueba</dt>
            <dd className="mt-2">
              Administrador: <span className="font-mono">admin@upse.edu.ec / 123456</span>
            </dd>
            <dd className="mt-1">
              Cliente: <span className="font-mono">cliente@upse.edu.ec / 123456</span>
            </dd>
          </dl>
        </div>
      </main>
    </div>
  );
};

export default Login;
