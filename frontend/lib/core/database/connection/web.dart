import 'package:drift/drift.dart';
import 'package:drift/wasm.dart';

QueryExecutor connect() {
  return LazyDatabase(() async {
    // Use WebWorkerConnection for better performance
    // This requires drift_worker.js to be generated
    final result = await WasmDatabase.open(
      databaseName: 'telegram_go_db',
      sqlite3Uri: Uri.parse('sqlite3.wasm'),
      driftWorkerUri: Uri.parse('drift_worker.js'),
    );
    return result.resolvedExecutor;
  });
}
