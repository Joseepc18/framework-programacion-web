const Dashboard = () => {
  return (
    <div>
      <h1 className="text-2xl font-bold text-slate-800 mb-6">
        Resumen General
      </h1>
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div className="bg-white p-6 rounded-lg shadow-sm border border-slate-100">
          <p className="text-sm text-slate-500 uppercase font-semibold">
            Ventas Totales
          </p>
          <p className="text-3xl font-bold text-indigo-600">$12,450.00</p>
        </div>
        <div className="bg-white p-6 rounded-lg shadow-sm border border-slate-100">
          <p className="text-sm text-slate-500 uppercase font-semibold">
            Referidos Activos
          </p>
          <p className="text-3xl font-bold text-indigo-600">24</p>
        </div>
        <div className="bg-white p-6 rounded-lg shadow-sm border border-slate-100">
          <p className="text-sm text-slate-500 uppercase font-semibold">
            Nivel Actual
          </p>
          <p className="text-3xl font-bold text-indigo-600">Diamante</p>
        </div>
      </div>
    </div>
  );
};

export default Dashboard;
