set -e

CONTAINER_NAME=swift-api-test-runner

echo "Cleaning up old containers..."
docker-compose -f docker-compose.test.yml down -v

echo "Starting test container..."
docker-compose -f docker-compose.test.yml up --build --abort-on-container-exit

echo "Copying test results..."
docker cp $CONTAINER_NAME:/app/test-report.txt ./test-report.txt || echo "test-report.txt not found"
docker cp $CONTAINER_NAME:/app/coverage.html ./coverage.html || echo "test coverage not found"

echo "Cleaning up..."
docker-compose -f docker-compose.test.yml down -v

echo "Finished!"
echo "Test report saved in: ./test-report.txt"
echo "Test coverage saved in: ./coverage.html"
