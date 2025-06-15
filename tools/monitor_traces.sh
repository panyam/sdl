#!/bin/bash

# SDL Measurement Monitor
# This script provides a simple way to monitor measurements in DuckDB

DB_PATH="/tmp/sdl_metrics/traces.duckdb"

if [ ! -f "$DB_PATH" ]; then
    echo "Database not found at: $DB_PATH"
    echo "Make sure you've started measurements in SDL console first."
    exit 1
fi

echo "SDL Measurement Monitor"
echo "Database: $DB_PATH"
echo ""
echo "Commands:"
echo "  1) Show recent traces (last 20)"
echo "  2) Show summary by target"
echo "  3) Show traces per second"
echo "  4) Live monitor (updates every 2s)"
echo "  5) Open interactive DuckDB shell"
echo ""
read -p "Select option (1-5): " choice

case $choice in
    1)
        duckdb -readonly $DB_PATH -c "
        SELECT 
            datetime(timestamp/1e9, 'unixepoch') as time,
            target,
            duration,
            run_id
        FROM traces
        ORDER BY timestamp DESC
        LIMIT 20;"
        ;;
    2)
        duckdb -readonly $DB_PATH -c "
        SELECT 
            target,
            COUNT(*) as count,
            ROUND(AVG(duration), 3) as avg_duration,
            ROUND(MIN(duration), 3) as min_duration,
            ROUND(MAX(duration), 3) as max_duration,
            ROUND(PERCENTILE_CONT(0.50) WITHIN GROUP (ORDER BY duration), 3) as p50,
            ROUND(PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY duration), 3) as p95
        FROM traces
        GROUP BY target;"
        ;;
    3)
        duckdb -readonly $DB_PATH -c "
        SELECT 
            target,
            strftime('%Y-%m-%d %H:%M:%S', datetime(timestamp/1e9, 'unixepoch')) as second,
            COUNT(*) as requests_per_second
        FROM traces
        WHERE timestamp > (unixepoch() - 60) * 1e9
        GROUP BY target, second
        ORDER BY second DESC
        LIMIT 20;"
        ;;
    4)
        echo "Live monitoring (Ctrl+C to stop)..."
        while true; do
            clear
            echo "=== SDL Measurements - $(date) ==="
            echo ""
            duckdb -readonly $DB_PATH -c "
            SELECT 
                datetime(timestamp/1e9, 'unixepoch') as time,
                target,
                duration,
                run_id
            FROM traces
            WHERE timestamp > (unixepoch() - 5) * 1e9
            ORDER BY timestamp DESC
            LIMIT 10;" 2>/dev/null
            
            echo ""
            echo "=== Summary (last 30s) ==="
            duckdb -readonly $DB_PATH -c "
            SELECT 
                target,
                COUNT(*) as count,
                ROUND(AVG(duration), 3) as avg_duration
            FROM traces
            WHERE timestamp > (unixepoch() - 30) * 1e9
            GROUP BY target;" 2>/dev/null
            
            sleep 2
        done
        ;;
    5)
        echo "Opening DuckDB shell (read-only mode)..."
        echo "Try: SELECT * FROM traces LIMIT 10;"
        duckdb -readonly $DB_PATH
        ;;
    *)
        echo "Invalid option"
        exit 1
        ;;
esac