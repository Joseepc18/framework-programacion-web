import { BrowserRouter, Route, Routes } from 'react-router-dom';
import Layout from './components/Layout';
import RutaProtegida from './components/RutaProtegida';
import { AREAS, rutaDeArea } from './config/permisos';
import { AuthProvider } from './context/AuthContext';
import AreaRolPage from './pages/AreaRolPage';
import CatalogosPage from './pages/CatalogosPage';
import HealthPage from './pages/HealthPage';
import HomePage from './pages/HomePage';
import LoginPage from './pages/LoginPage';
import NotFoundPage from './pages/NotFoundPage';
import PacienteFichaPage from './pages/PacienteFichaPage';
import PacienteFormPage from './pages/PacienteFormPage';
import PacientesPage from './pages/PacientesPage';

const App = () => (
  <AuthProvider>
    <BrowserRouter>
      <Routes>
        <Route path="login" element={<LoginPage />} />
        <Route element={<RutaProtegida />}>
          <Route element={<Layout />}>
            <Route index element={<HomePage />} />
            <Route element={<RutaProtegida roles={['admin_ti']} />}>
              <Route path="estado" element={<HealthPage />} />
              <Route path="catalogos" element={<CatalogosPage />} />
            </Route>
            <Route element={<RutaProtegida roles={['recepcionista', 'odontologo', 'asistente']} />}>
              <Route path="pacientes" element={<PacientesPage />} />
              <Route path="pacientes/:id" element={<PacienteFichaPage />} />
            </Route>
            <Route element={<RutaProtegida roles={['recepcionista']} />}>
              <Route path="pacientes/nuevo" element={<PacienteFormPage />} />
              <Route path="pacientes/:id/editar" element={<PacienteFormPage />} />
            </Route>
            {AREAS.map((area) => (
              <Route key={area.slug} element={<RutaProtegida roles={[area.rol]} />}>
                <Route path={rutaDeArea(area).slice(1)} element={<AreaRolPage area={area} />} />
              </Route>
            ))}
            <Route path="*" element={<NotFoundPage />} />
          </Route>
        </Route>
      </Routes>
    </BrowserRouter>
  </AuthProvider>
);

export default App;
