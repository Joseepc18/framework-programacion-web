// src/components/Login.tsx
import { useState, type FormEvent } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useAuth, type Rol } from "../context/AuthContext";
import { apiFetch } from "../services/http";

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

  return (
    <div className="min-h-screen flex items-center justify-center bg-slate-100 px-4">
      <div className="max-w-md w-full bg-white rounded-xl shadow-md p-8 border border-slate-200">
        <div className="text-center mb-8">
          <h2 className="text-3xl font-bold text-slate-900">MultiCatálogo</h2>
          <p className="text-slate-500 mt-2">
            Ingresa a tu cuenta para continuar
          </p>
        </div>

        {error && (
          <div className="bg-red-50 text-red-600 p-3 rounded-lg text-sm mb-6 text-center border border-red-200">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-6">
          <div>
            <label className="block text-sm font-medium text-slate-700 mb-2">
              Correo Electrónico
            </label>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="w-full px-4 py-3 rounded-lg border border-slate-300 focus:ring-2 focus:ring-indigo-600 focus:border-indigo-600 outline-none transition"
              placeholder="admin@upse.edu.ec"
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
              required
            />
          </div>

          <button
            type="submit"
            disabled={loading}
            className="w-full bg-indigo-600 text-white font-bold py-3 rounded-lg hover:bg-indigo-700 transition disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {loading ? "Validando..." : "Iniciar Sesión"}
          </button>
        </form>

        <p className="text-center text-sm text-slate-500 mt-6">
          ¿Aún no tienes cuenta?{" "}
          <Link to="/registro" className="text-indigo-600 font-semibold hover:underline">
            Crear cuenta
          </Link>
        </p>

        <div className="mt-6 p-4 bg-slate-50 rounded-lg border border-slate-200 text-xs text-slate-600 space-y-1">
          <p className="font-semibold text-slate-700">Cuentas de prueba:</p>
          <p>👑 Admin: <span className="font-mono">admin@upse.edu.ec / 123456</span></p>
          <p>🛍️ Cliente: <span className="font-mono">cliente@upse.edu.ec / 123456</span></p>
        </div>
      </div>
    </div>
  );
};

export default Login;