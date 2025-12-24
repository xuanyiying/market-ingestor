import React, { useMemo, useRef, useEffect } from 'react';
import ReactECharts from 'echarts-for-react';
import { useMarketStore } from '../store/useMarketStore';

const upColor = '#00c087';
const downColor = '#ff3b30';

const Chart: React.FC = () => {
  const chartRef = useRef<ReactECharts>(null);
  const { klines } = useMarketStore();

  const options = useMemo(() => {
    const dates = klines.map(k => {
      const d = new Date(k.t);
      return d.toLocaleDateString() + ' ' + d.toLocaleTimeString();
    });

    const candles = klines.map(k => [
      parseFloat(k.o),
      parseFloat(k.c),
      parseFloat(k.l),
      parseFloat(k.h)
    ]);

    const volumes = klines.map(k => parseFloat(k.v));

    return {
      backgroundColor: '#1e1e1e',
      animation: false,
      legend: {
        bottom: 10,
        left: 'center',
        data: ['K-Line', 'Volume'],
        textStyle: { color: '#ccc' }
      },
      tooltip: {
        trigger: 'axis',
        axisPointer: { type: 'cross' },
        backgroundColor: 'rgba(50, 50, 50, 0.9)',
        borderColor: '#444',
        textStyle: { color: '#eee' },
      },
      grid: [
        { left: '50', right: '50', height: '60%' },
        { left: '50', right: '50', top: '75%', height: '15%' }
      ],
      xAxis: [
        {
          type: 'category',
          data: dates,
          boundaryGap: false,
          axisLine: { onZero: false, lineStyle: { color: '#444' } },
          splitLine: { show: false },
          min: 'dataMin',
          max: 'dataMax',
        },
        {
          type: 'category',
          gridIndex: 1,
          data: dates,
          boundaryGap: false,
          axisLine: { onZero: false, lineStyle: { color: '#444' } },
          axisTick: { show: false },
          splitLine: { show: false },
          axisLabel: { show: false },
          min: 'dataMin',
          max: 'dataMax'
        }
      ],
      yAxis: [
        {
          scale: true,
          splitArea: { show: false },
          axisLine: { lineStyle: { color: '#444' } },
          splitLine: { lineStyle: { color: '#333' } },
        },
        {
          scale: true,
          gridIndex: 1,
          splitNumber: 2,
          axisLabel: { show: false },
          axisLine: { show: false },
          axisTick: { show: false },
          splitLine: { show: false }
        }
      ],
      dataZoom: [
        {
          type: 'inside',
          xAxisIndex: [0, 1],
          start: 50,
          end: 100
        },
        {
          show: true,
          xAxisIndex: [0, 1],
          type: 'slider',
          top: '92%',
          start: 50,
          end: 100,
          textStyle: { color: '#888' }
        }
      ],
      series: [
        {
          name: 'K-Line',
          type: 'candlestick',
          data: candles,
          itemStyle: {
            color: upColor,
            color0: downColor,
            borderColor: upColor,
            borderColor0: downColor
          }
        },
        {
          name: 'Volume',
          type: 'bar',
          xAxisIndex: 1,
          yAxisIndex: 1,
          data: volumes,
          itemStyle: {
            color: (params: any) => {
              const idx = params.dataIndex;
              if (idx === 0) return upColor;
              return candles[idx][1] >= candles[idx][0] ? upColor : downColor;
            }
          }
        }
      ]
    };
  }, [klines]);

  useEffect(() => {
    if (chartRef.current) {
      const chartInstance = chartRef.current.getEchartsInstance();
      chartInstance.resize();
    }
  }, []);

  return (
    <div className="w-full h-[600px]">
      <ReactECharts
        ref={chartRef}
        option={options}
        style={{ height: '100%', width: '100%' }}
        theme="dark"
        notMerge={true}
        lazyUpdate={true}
      />
    </div>
  );
};

export default Chart;
