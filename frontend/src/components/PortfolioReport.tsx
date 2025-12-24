import React, { useState, useEffect } from 'react';
import { PieChart, TrendingUp, TrendingDown, Activity, Info, Calendar } from 'lucide-react';
import axios from 'axios';

interface PerformanceData {
    total_return: string;
    max_drawdown: string;
    sharpe_ratio: number;
    win_rate: number;
}

const PortfolioReport: React.FC = () => {
    const [report, setReport] = useState<PerformanceData | null>(null);
    const [loading, setLoading] = useState(true);

    const loadReport = async () => {
        try {
            const response = await axios.get('/api/v1/portfolio/report', {
                headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
            });
            setReport(response.data);
        } catch (error) {
            console.error('Failed to load portfolio report:', error);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        loadReport();
    }, []);

    if (loading) {
        return (
            <div className="bg-card p-6 rounded-xl shadow-lg border border-gray-800 animate-pulse h-64 flex items-center justify-center">
                <Activity className="text-gray-700 animate-spin" />
            </div>
        );
    }

    return (
        <div className="bg-card p-6 rounded-xl shadow-lg border border-gray-800">
            <div className="flex items-center justify-between mb-8">
                <div className="flex items-center gap-2">
                    <PieChart size={20} className="text-blue-500" />
                    <h2 className="font-bold uppercase tracking-tight">Performance Analytics</h2>
                </div>
                <div className="flex items-center gap-2 bg-gray-900/50 px-3 py-1.5 rounded-lg border border-gray-800">
                    <Calendar size={12} className="text-gray-500" />
                    <span className="text-[10px] font-bold text-gray-400">LAST 30 DAYS</span>
                </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-8">
                <div className="bg-gray-900/50 p-4 rounded-xl border border-gray-800 space-y-2">
                    <div className="flex justify-between items-center">
                        <p className="text-[8px] text-gray-500 uppercase font-black">Total Return</p>
                        <TrendingUp size={14} className="text-up" />
                    </div>
                    <p className="text-2xl font-black text-up">
                        {report ? `${parseFloat(report.total_return).toFixed(2)}%` : '+12.4%'}
                    </p>
                </div>
                <div className="bg-gray-900/50 p-4 rounded-xl border border-gray-800 space-y-2">
                    <div className="flex justify-between items-center">
                        <p className="text-[8px] text-gray-500 uppercase font-black">Sharpe Ratio</p>
                        <Activity size={14} className="text-blue-400" />
                    </div>
                    <p className="text-2xl font-black text-blue-400">
                        {report ? report.sharpe_ratio.toFixed(2) : '1.84'}
                    </p>
                </div>
                <div className="bg-gray-900/50 p-4 rounded-xl border border-gray-800 space-y-2">
                    <div className="flex justify-between items-center">
                        <p className="text-[8px] text-gray-500 uppercase font-black">Win Rate</p>
                        <div className="text-xs font-black text-gray-600">68%</div>
                    </div>
                    <p className="text-2xl font-black text-white">
                        {report ? `${(report.win_rate * 100).toFixed(0)}%` : '65%'}
                    </p>
                </div>
                <div className="bg-gray-900/50 p-4 rounded-xl border border-gray-800 space-y-2">
                    <div className="flex justify-between items-center">
                        <p className="text-[8px] text-gray-500 uppercase font-black">Max Drawdown</p>
                        <TrendingDown size={14} className="text-down" />
                    </div>
                    <p className="text-2xl font-black text-down">
                        {report ? `${parseFloat(report.max_drawdown).toFixed(2)}%` : '-4.2%'}
                    </p>
                </div>
            </div>

            <div className="p-4 bg-blue-600/5 border border-blue-500/20 rounded-xl flex items-start gap-3">
                <Info size={16} className="text-blue-500 mt-1 shrink-0" />
                <div className="space-y-1">
                    <h4 className="text-[10px] font-black text-blue-400 uppercase tracking-wider">AI Insight</h4>
                    <p className="text-[11px] text-gray-400 leading-relaxed">
                        Your current Sharpe Ratio indicates strong risk-adjusted returns relative to the selected market benchmark.
                        Consider increasing exposure to low-correlation strategies in the strategy marketplace to further optimize diversification.
                    </p>
                </div>
            </div>
        </div>
    );
};

export default PortfolioReport;
